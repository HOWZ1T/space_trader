package space_trader

import (
	"fmt"
	"github.com/HOWZ1T/space_trader/assert"
	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"math"
	"os"
	"testing"
)

var username = ""
var token = ""
var stTest *SpaceTrader

func setup() {
	err := godotenv.Load(".test_env")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("ST_LOG", "verbose")
	if err != nil {
		panic(err)
	}

	username = os.Getenv("ST_USERNAME")
	token = os.Getenv("ST_TOKEN")
	stTest = New(token, username)

}

func teardown() {}

func TestMain(m *testing.M) {
	setup()
	stTest.disallowUnknownFields = true
	code := m.Run()
	teardown()
	os.Exit(code)
}

func createUserAndTakeLoanAndBuyShip(st *SpaceTrader) (models.Account, error) {
	// create and switch to user
	uid, err := uuid.NewUUID()
	if err != nil {
		return models.Account{}, err
	}

	uname := uid.String()
	token, err := st.RegisterUser(uname)
	if err != nil {
		return models.Account{}, err
	}

	st.SwitchUser(token, uname)

	// take out loan
	acc, err := st.TakeLoan("startup")
	if err != nil {
		return models.Account{}, err
	}

	ships, err := st.AvailableShips("")
	if err != nil {
		return models.Account{}, err
	}

	// buy cheapest ship
	cheapestShip := struct {
		shipType string
		location string
	}{}
	min := models.Currency(math.MaxInt32)
	for _, ship := range ships {
		for _, ploc := range ship.PurchaseLocation {
			if ploc.Price <= min {
				cheapestShip.shipType = ship.Type
				cheapestShip.location = ploc.Location
				min = ploc.Price
			}
		}
	}

	acc, err = st.BuyShip(cheapestShip.location, cheapestShip.shipType)
	if err != nil {
		return models.Account{}, err
	}

	return acc, nil
}

func TestApiStatus(t *testing.T) {
	stat, err := stTest.ApiStatus()
	(*assert.T)(t).Nil(err)

	expected := "spacetraders is currently online and available to play"
	(*assert.T)(t).Equals(expected, stat)
}

func TestRegisterUser(t *testing.T) {
	username := uuid.New().String()
	s, err := stTest.RegisterUser(username)
	(*assert.T)(t).Nil(err)

	if len(s) <= 0 {
		t.Errorf("expected token, got: %s", s)
	}

	// test event
	select {
	case evt := <-stTest.EventsChannel():
		if e, ok := evt.(*events.UserRegistered); ok {
			(*assert.T)(t).Equals(e.Token, s)
			(*assert.T)(t).Equals(e.Username, username)
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

func TestAccount(t *testing.T) {
	acc, err := stTest.Account()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(username, acc.Username)
}

func TestAccountError(t *testing.T) {
	uid := uuid.New().String()
	st := New(token, uid)
	st.disallowUnknownFields = true
	_, err := st.Account()
	(*assert.T)(t).NotNil(err)

	if e, ok := err.(*errs.ApiError); ok {
		expected := fmt.Sprintf("[404] error - User %s does not exist!", uid)
		got := e.Error()
		(*assert.T)(t).Equals(expected, got)
	}
}

func TestAvailableLoans(t *testing.T) {
	loans, err := stTest.AvailableLoans()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(loans[0].Type, "STARTUP")
}

func TestMyLoans(t *testing.T) {
	loans, err := stTest.MyLoans()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(loans[0].Type, "STARTUP")
}

func TestTakeLoan(t *testing.T) {
	username := uuid.New().String()
	token, err := stTest.RegisterUser(username)
	(*assert.T)(t).Nil(err)

	st := New(token, username)
	st.disallowUnknownFields = true
	acc, err := st.TakeLoan("startup")

	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(acc)
	(*assert.T)(t).Equals(len(acc.Loans) > 0, true)

	// test event
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.Loan); ok {
			(*assert.T)(t).Equals(len(e.Account.Loans) > 0, true)
			(*assert.T)(t).Equals(e.Type, "PURCHASED")
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

func TestAvailableShips(t *testing.T) {
	ships, err := stTest.AvailableShips("")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(ships)
	(*assert.T)(t).Equals(len(ships) > 0, true)
}

func TestAvailableShipsFiltered(t *testing.T) {
	ships, err := stTest.AvailableShips("MK-II")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(ships)
	(*assert.T)(t).Equals(len(ships) > 0, true)

	for _, ship := range ships {
		(*assert.T)(t).Equals(ship.Class, "MK-II")
	}
}

func TestBuyShipNoFunds(t *testing.T) {
	username := uuid.New().String()
	token, err := stTest.RegisterUser(username)
	(*assert.T)(t).Nil(err)

	st := New(token, username)
	st.disallowUnknownFields = true
	_, err = st.BuyShip("OE-PM-TR", "JW-MK-I")

	(*assert.T)(t).NotNil(err)
	(*assert.T)(t).Equals("[400] error - User has insufficient funds to purchase ship.", err.Error())
}

func TestBuyShipWithFunds(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	(*assert.T)(t).NotNil(acc.Ships)
	(*assert.T)(t).NotEquals(len(acc.Ships), 0)
	(*assert.T)(t).Equals(acc.Ships[0].Type, "JW-MK-I")

	// test event
	// clear register, user switch, & loan events
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.ShipPurchased); ok {
			(*assert.T)(t).Equals(e.Account.Ships[0].Type, "JW-MK-I")
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

func TestBuyGoods(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	shipID := acc.Ships[0].ID
	order, err := st.BuyGood(shipID, "FUEL", 1)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders), 0)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")

	// test event
	// clear register, user switch, loan & ship purchased events
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.ShipOrder); ok {
			(*assert.T)(t).Equals(e.Type, "BUY")
			(*assert.T)(t).Equals(e.Order.Orders[0].Good, "FUEL")
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

func TestSellGoods(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	shipID := acc.Ships[0].ID
	order, err := st.BuyGood(shipID, "FUEL", 1)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders), 0)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")

	order, err = st.SellGood(shipID, "FUEL", 1)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders), 0)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")

	// test event
	// clear register, user switch, loan, ship purchased, & buy good events
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.ShipOrder); ok {
			(*assert.T)(t).Equals(e.Type, "SELL")
			(*assert.T)(t).Equals(e.Order.Orders[0].Good, "FUEL")
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

func TestSearchSystem(t *testing.T) {
	locs, err := stTest.SearchSystem("OE", "ASTEROID")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(locs), 0)
	(*assert.T)(t).Equals(locs[0].Type, "ASTEROID")
}

func TestFlightPlan(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	shipID := acc.Ships[0].ID
	order, err := st.BuyGood(shipID, "FUEL", 80)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders)+1, 1)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")

	plan, err := st.CreateFlightPlan(shipID, "OE-CR")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(plan)

	plan2, err := st.GetFlightPlan(plan.ID)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(plan2)
	(*assert.T)(t).Equals(plan.ID, plan2.ID)

	(*assert.T)(t).Equals(st.flightPlans[plan.ID].Destination, plan.Destination)

	// test event
	// clear register, user switch, loan, ship purchased, & buy good events
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	_ = <-st.EventsChannel()
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.FlightPlan); ok {
			(*assert.T)(t).Equals(e.Type, "CREATED")
			(*assert.T)(t).Equals(e.Plan.ID, plan.ID)
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}

	// Get all flight plans in system and confirm our plan is one of them
	flightPlans, err := st.GetAllFlightPlansWithinSystem("OE")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(flightPlans), 0)
	foundShipIDInPlans := false
	for _, flightPlan := range flightPlans {
		if flightPlan.ShipID == shipID {
			foundShipIDInPlans = true
		}
	}
	(*assert.T)(t).Equals(true, foundShipIDInPlans)
}

func TestPayLoan(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	_, err = st.PayLoan(acc.Loans[0].ID)
	(*assert.T)(t).NotNil(err)
	(*assert.T)(t).Equals("[400] error - Insufficient funds to pay for loan.", err.Error())
}

func TestGetLocation(t *testing.T) {
	loc, err := stTest.GetLocation("OE-CR")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(loc.Name, "Carth")
	(*assert.T)(t).Equals(loc.Type, "PLANET")
}

func TestGetLocationsInSystem(t *testing.T) {
	locs, err := stTest.GetLocationsInSystem("OE")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(locs)
	(*assert.T)(t).NotEquals(len(locs), 0)
	(*assert.T)(t).Equals(locs[0].Type, "PLANET")
}

func TestGetMarket(t *testing.T) {
	st := New("", "")
	st.disallowUnknownFields = true
	acc, err := createUserAndTakeLoanAndBuyShip(st)
	(*assert.T)(t).Nil(err)

	locSymbol := acc.Ships[0].Location
	market, err := st.GetMarket(locSymbol)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(market.Goods)
	(*assert.T)(t).NotEquals(len(market.Goods), 0)
	(*assert.T)(t).NotEquals(len(market.Goods[0].Symbol), 0)
}

func Test_GetSystems(t *testing.T) {
	systems, err := stTest.GetSystems()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(systems)
	(*assert.T)(t).NotEquals(len(systems), 0)
	(*assert.T)(t).Equals(systems[0].Symbol, "OE")
	(*assert.T)(t).Equals(systems[0].Locations[0].Type, "PLANET")
}

func Test_SwitchUser(t *testing.T) {
	st := New("a", "b")
	st.disallowUnknownFields = true
	(*assert.T)(t).Equals(st.token, "a")
	(*assert.T)(t).Equals(st.username, "b")

	st.SwitchUser("c", "d")

	(*assert.T)(t).Equals(st.token, "c")
	(*assert.T)(t).Equals(st.username, "d")

	// test event
	select {
	case evt := <-st.EventsChannel():
		if e, ok := evt.(*events.UserSwitched); ok {
			(*assert.T)(t).Equals(e.Token, "c")
			(*assert.T)(t).Equals(e.Username, "d")
		} else {
			(*assert.T)(t).Errorf("event is not of the proper type")
		}
		break

	default:
		(*assert.T)(t).Errorf("event is not of the proper type")
		break
	}
}

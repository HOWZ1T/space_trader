package space_trader

import (
	"fmt"
	"github.com/HOWZ1T/space_trader/assert"
	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"math"
	"os"
	"testing"
)

var username = ""
var token = ""
var stTest SpaceTrader

func setup() {
	err := godotenv.Load(".test_env")
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
	code := m.Run()
	teardown()
	os.Exit(code)
}

func createUserAndTakeLoanAndBuyShip(st *SpaceTrader) (models.Account, error) {
	// create and switch to user
	uname := uuid.NewString()
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
}

func TestAccount(t *testing.T) {
	acc, err := stTest.Account()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(username, acc.Username)
}

func TestAccountError(t *testing.T) {
	uid := uuid.New().String()
	st := New(token, uid)
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
	acc, err := st.TakeLoan("startup")

	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(acc)
	(*assert.T)(t).Equals(len(acc.Loans) > 0, true)
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
}

func TestBuyShipNoFunds(t *testing.T) {
	username := uuid.New().String()
	token, err := stTest.RegisterUser(username)
	(*assert.T)(t).Nil(err)

	st := New(token, username)
	_, err = st.BuyShip("OE-G4", "JW-MK-I")

	(*assert.T)(t).NotNil(err)
	(*assert.T)(t).Equals("[400] error - User has insufficient funds to purchase ship.", err.Error())
}

func TestBuyShipWithFunds(t *testing.T) {
	st := New("", "")
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
	(*assert.T)(t).Nil(err)

	(*assert.T)(t).NotNil(acc.Ships)
	(*assert.T)(t).NotEquals(len(acc.Ships), 0)
	(*assert.T)(t).Equals(acc.Ships[0].Type, "JW-MK-I")
}

func TestBuyGoods(t *testing.T) {
	st := New("", "")
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
	(*assert.T)(t).Nil(err)

	shipID := acc.Ships[0].ID
	order, err := st.BuyGood(shipID, "FUEL", 1)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders), 0)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")
}

func TestSellGoods(t *testing.T) {
	st := New("", "")
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
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
}

func TestSearchSystem(t *testing.T) {
	locs, err := stTest.SearchSystem("OE", "ASTEROID")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(locs), 0)
	(*assert.T)(t).Equals(locs[0].Type, "ASTEROID")
}

func TestFlightPlan(t *testing.T) {
	st := New("", "")
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
	(*assert.T)(t).Nil(err)

	shipID := acc.Ships[0].ID
	order, err := st.BuyGood(shipID, "FUEL", 80)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotEquals(len(order.Orders)+1, 1)
	(*assert.T)(t).Equals(order.Orders[0].Good, "FUEL")

	plan, err := st.CreateFlightPlan(shipID, "OE-ZEP4")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(plan)

	plan2, err := st.GetFlightPlan(plan.ID)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(plan2)
	(*assert.T)(t).Equals(plan.ID, plan2.ID)
}

func TestPayLoan(t *testing.T) {
	st := New("", "")
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
	(*assert.T)(t).Nil(err)

	_, err = st.PayLoan(acc.Loans[0].ID)
	(*assert.T)(t).NotNil(err)
	(*assert.T)(t).Equals("[400] error - Insufficient funds to pay for loan.", err.Error())
}

func TestGetLocation(t *testing.T) {
	loc, err := stTest.GetLocation("OE-D2")
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).Equals(loc.Name, "Delta II")
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
	acc, err := createUserAndTakeLoanAndBuyShip(&st)
	(*assert.T)(t).Nil(err)

	locSymbol := acc.Ships[0].Location
	market, err := st.GetMarket(locSymbol)
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(market.Goods)
	(*assert.T)(t).NotEquals(len(market.Goods), 0)
	(*assert.T)(t).Equals(market.Goods[0].Symbol, "ELECTROINICS")
}

func Test_GetSystems(t *testing.T) {
	systems, err := stTest.GetSystems()
	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(systems)
	(*assert.T)(t).NotEquals(len(systems), 0)
	(*assert.T)(t).Equals(systems[0].Symbol, "OE")
	(*assert.T)(t).Equals(systems[0].Locations[0].Type, "PLANET")
}

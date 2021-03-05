package space_trader

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"os"
	"space_trader/assert"
	"space_trader/errs"
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
		expected := fmt.Sprintf("User %s does not exist!", uid)
		got := e.Error()
		(*assert.T)(t).Equals(expected, got)
	}
}

func TestAvailableLoans(t *testing.T) {
	loans, err := stTest.AvailableLoans()
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

func TestBuyShip(t *testing.T) {
	username := uuid.New().String()
	token, err := stTest.RegisterUser(username)
	(*assert.T)(t).Nil(err)

	st := New(token, username)
	acc, err := st.BuyShip("OE-G4", "JW-MK-I")

	(*assert.T)(t).Nil(err)
	(*assert.T)(t).NotNil(acc)
	fmt.Println(acc)
	(*assert.T)(t).Equals(len(acc.Ships) > 0, true)
}

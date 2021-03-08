package models

// Provides useful constants ('enums') for the project.
// Const structure OBJECT_ENUM, e.g.: LOAN_STARTUP
const (
	LOAN_STARTUP = "STARTUP"

	LOC_PLANET    = "PLANET"
	LOC_MOON      = "MOON"
	LOC_GAS_GIANT = "GAS_GIANT"
	LOC_ASTEROID  = "ASTEROID"

	GOOD_MACHINERY = "MACHINERY"
	GOOD_RESEARCH  = "RESEARCH"
	GOOD_CHEMICALS = "CHEMICALS"
	GOOD_FOOD      = "FOOD"
	GOOD_FUEL      = "FUEL"
	GOOD_WORKERS   = "WORKERS"

	SHIP_CLASS_I   = "MK-I"
	SHIP_CLASS_II  = "MK-II"
	SHIP_CLASS_III = "MK-III"
)

// vars for providing 'nil' struct definitions
var (
	NilAccount          = Account{}
	NilCargo            = Cargo{}
	NilFlightPlan       = FlightPlan{}
	NilLoan             = Loan{}
	NilPurchaseLocation = PurchaseLocation{}
	NilLocation         = Location{}
	NilMarketLocation   = MarketLocation{}
	NilGood             = Good{}
	NilMarket           = Market{}
	NilOrder            = Order{}
	NilShipOrder        = ShipOrder{}
	NilShip             = Ship{}
	NilSystem           = System{}
)

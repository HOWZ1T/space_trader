# 🚀 Events
All events will be time stamped and named.
These properties can be accessed through the `event.Time()` and `event.Name()` functions.

Events are streamed over a buffered channel which is accessible through 
```go
st := space_traders.New("TOKEN", "USERNAME").EventsChannel()
``` 
function.

## 🔧 Event Reference
The event reference is structured as follows:
```
[Event] Name of the event (event.Name())
  - Model, a link to the source for the event struct.
  - [Type] Name of type (event.Type)
    - Some events may have this field.
    - Sub type, e.g.: a ship order is either BUY or SELL
  - [Field] Name of public fields contained in this event.
    - e.g: For a flight plan event: 
           Plan is accessible through event.Plan
```
Events will often have fields containing a model with the relevant data.
E.g.: A FLIGHT_PLAN event will contain the `models.FlightPlan` structure under the `event.Plan` field.

- [Event] FLIGHT_PLAN
  - [Model](events/flight_plan.go)
  - [Type] CREATED
    - Triggered when a new flight plan is created.
  - [Type] ENDED
    - Triggered when a flight plan is finished, i.e.: the ship reaches its destination.
  - [Field] Plan (`event.Plan`)


- [Event] LOAN
  - [Model](events/loan.go)
  - [Type] PURCHASED
      - Triggered when a new loan is taken out.
  - [Type] PAID
      - Triggered when a loan has been paid.
  - [Field] Account (`event.Account`)


- [Event] SHIP_ORDER
  - [Model](events/ship_order.go)
  - [Type] BUY
      - Triggered when a ship buys goods.
  - [Type] SELL
      - Triggered when a ship sells goods.
  - [Field] Order (`event.Order`)


- [Event] SHIP_PURCHASED
  - [Model](events/ship_purchased.go)
  - Triggered when a user buys a ship.
  - [Field] Account (`event.Account`)   


- [Event] USER_REGISTERED
  - Triggered when a user is registered.
  - [Model](events/user_registered.go)
  - [Field] Username (`event.Username`)
  - [Field] Token (`event.Token`)
    

- [Event] USER_SWITCHED
  - Triggered when the wrapper switches user.
  - [Model](events/user_switched.go)
  - [Field] Username (`event.Username`)
  - [Field] Token (`event.Token`)
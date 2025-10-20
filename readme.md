
client booking journey: (bookingService service)
1) the client enters the site and see details of all cottages; (GET /cottages)
    1.1) the client can check one specific cottage and see more details; (GET /cottage/:name)
    1.2) the client can search for availabilities; (GET /cottage/:name/available-dates)
    1.3) and book it right away (POST /cottage/:name/booking)
2) the client can search for availability for a given cottage or for a given cottage type; (GET /cottage/:name/available-dates GET /cottage/type/:cottageType/available-dates)
3) the client can select a period and ask for a booking; (POST cottage/:name/booking)
4) a message of payment required will be sent to the client; (finance service NOT IMPLEMENTED YET)
5) after confirmation of payment, the booking confirmation will be sent; (finance service NOT IMPLEMENTED YET)
6*) the client can update of delete a booking if he/she wants, further fees my apply; (PATCH /cottage/:name/booking/:bookingId DELETE /cottage/:name/booking/:bookingId)

** staffManger service needs to manage the following (separate in its own micro service even if small)
    - receptionManager service
    - restaurantManager service
    - cleaningManager service
    - supplyManager service

** financeManager service needs to:
    1) generate financial reports;
    2) process booking payment;
    3) order new supplies;

client stay journey: (operationManager service)
1) the client checks in the Villa
2) the client receives the key
3) staff prepares the bath and the cottage for sleeping
4) staff prepares the breakfast
5) staff cleans the cottage

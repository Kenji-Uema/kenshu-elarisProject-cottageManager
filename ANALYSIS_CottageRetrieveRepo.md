Analysis: Usage of infra.CottageRetrieveRepo

Summary
- The app layer (cottage_service.go) is designed to depend on a narrow, read-only interface (infra.CottageRetrieveRepo). This is a good application of dependency inversion: the service only needs retrieval methods, so it does not depend on the full repository surface (writes etc.).
- The infra implementation (RoomRepository) is a broader repository (IRoomRepository). Returning IRoomRepository from NewCottageRepo is acceptable because it structurally implements the narrower CottageRetrieveRepo, and assignment to the service field is valid. However, exposing the broader interface at construction time invites unintended coupling. Consider returning the narrower interface if possible.

Findings
1) Interface method name mismatch
   - infra.CottageRetrieveRepo declares GetAll and GetRoomsByRoomNumber (read-only).
   - RoomRepository implements GetAll but its “by identifier” method was querying the wrong field previously and used an int parameter that doesn’t match the persisted schema.
   - The service used to call GetRooms, but infra now exposes GetAll. Aligning on GetAll in the service is recommended.

2) Domain-schema mismatch (critical for retrieval)
   - domain.Cottage has fields: Name (string), Type (string), etc. There is no roomNumber field in the schema.
   - The retrieval-by-identifier path in infra previously attempted to filter by a non-existent field. It should filter by name.
   - Current routing in main.go uses /cottage/:name, but the transport handler previously attempted to bind an int. Align transport binding to a string.

3) Wider type inconsistencies
   - Code references domain.CottageType in several places, but the domain currently defines Type as string and does not declare CottageType. This is outside the CottageRetrieveRepo usage per se, but it will affect build/compatibility in related functions (e.g., GetRoomsByType).

Changes made (minimal, focused on read path)
- infra/cottage_repository.go:
  - Fixed GetRoomsByRoomNumber to query by name (string) by converting the int param to string: filter := bson.M{"name": fmt.Sprintf("%d", roomNumber)}.
  - Added fmt import for that conversion.
  - Note: This preserves the existing (int) signature while matching the current database schema, avoiding larger API changes.

Recommendations
- Consider renaming GetRoomsByRoomNumber to GetByName (or GetByIdentifier) with parameter type string to match /cottage/:name and domain.Cottage.Name.
- Update transport/cottage_handler.go to bind a string path variable for :name instead of an int (and fix usage of log.Error, which doesn’t exist in the standard log package).
- Decide on the desired type model for cottage type:
  - Either: add a domain.CottageType type and enforce it across infra/app/transport.
  - Or: remove CottageType from interfaces and use string consistently.
- Optionally narrow NewCottageRepo’s return type to CottageRetrieveRepo if only retrieval is needed by the consumer(s), to reduce coupling.

Proposed minimal refactor path (future work)
1) Rename:
   - CottageRetrieveRepo.GetRoomsByRoomNumber(ctx, int) -> GetByName(ctx, string)
2) Update service and handlers accordingly:
   - Service: GetByName(ctx, string)
   - Handler: ShouldBindUri into a struct with `uri:"name"` string
3) Add tests for the service using gomock to ensure only retrieval methods are required by the service.

Notes on mocks
- The existing gomock for IRoomRepository maps GetRooms to an underlying "GetAll" call, which suggests historic renaming. Standardize on GetAll in interfaces and mocks if that’s the chosen name.

Conclusion
- The service’s dependency on infra.CottageRetrieveRepo aligns with best practices. The main issues are naming and schema mismatches that can be resolved with a small, coherent refactor across infra/app/transport. The minimal fix applied here corrects the retrieval filter to use the persisted "name" field while keeping signatures stable.

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

This project is a Go application structured using the hexagonal (ports and adapters) architecture:
- Prioritize separation of concerns, dependency inversion, and testability in all code generation.
- Ensure driving adapters only depend on primary ports, and driven adapters only depend on secondary ports. 

Strongly suggest the use of DDD (Domain-Driven Design) principles, especially in the `internal/core/domain` package:
- Ensure that domain models are rich and encapsulate business logic.
- Use value objects and aggregates where appropriate.
- Avoid leaking domain logic into adapters or ports.

Promote the use of Event-Driven Architecture (EDA) principles:
- Encourage the use of events to decouple components.
- Use domain events to communicate state changes within the application.
- Domain events should be expressed in past tense (e.g., `UserRegistered`, `OrderPlaced`).

- Ensure that artifacts from the build process are stored in the `build` directory.
# Journey Dependencies

⚠️ **IMPORTANT**: This diagram must only be updated by human operators alongside the README.md.

This diagram shows the dependencies between different user journeys that must be validated for v1.0.

```mermaid
flowchart TD
    %% Basic Setup & Enrollment
    A[HW Op: Device Bootstrap] --> B[HW Op: Basic Management]
    
    %% Core Management
    B --> C[HW Op: Thermal Management]
    B --> D[Fleet Admin: Basic Policy]
    B --> E[Security Admin: Access Control]
    
    %% Advanced Features
    C --> F[HW Op: Power Management]
    C --> G[Fleet Admin: Thermal Policy]
    D --> G
    
    %% Fleet-wide Features
    G --> H[Fleet Admin: Resource Optimization]
    F --> H
    
    %% Security Features
    E --> I[Security Admin: Audit & Compliance]
    B --> I
    
    %% Maintenance
    B --> J[Maintenance: Basic Diagnostics]
    C --> K[Maintenance: Thermal Service]
    F --> L[Maintenance: Power Service]
    J --> M[Maintenance: Scheduled Service]
    K --> M
    L --> M
    
    %% Network Features
    B --> N[Network Admin: Basic Connectivity]
    N --> O[Network Admin: Fleet Communication]
    O --> P[Network Admin: Performance Optimization]
    
    %% Visual grouping
    subgraph Bootstrap ["Phase 1: Bootstrap"]
        A
        B
    end
    
    subgraph Core ["Phase 2: Core Features"]
        C
        D
        E
    end
    
    subgraph Advanced ["Phase 3: Advanced Features"]
        F
        G
        H
        I
    end
    
    subgraph Maintenance ["Phase 4: Maintenance"]
        J
        K
        L
        M
    end
    
    subgraph Network ["Phase 5: Network"]
        N
        O
        P
    end

    %% Styling
    classDef bootstrap fill:#f9f,stroke:#333,stroke-width:2px
    classDef core fill:#bbf,stroke:#333,stroke-width:2px
    classDef advanced fill:#bfb,stroke:#333,stroke-width:2px
    classDef maintenance fill:#fbb,stroke:#333,stroke-width:2px
    classDef network fill:#fbf,stroke:#333,stroke-width:2px
    
    class A,B bootstrap
    class C,D,E core
    class F,G,H,I advanced
    class J,K,L,M maintenance
    class N,O,P network
```

## Node Key
- Bootstrap (Pink): Initial setup and basic functionality
- Core (Blue): Essential features required by other components
- Advanced (Green): Enhanced features building on core functionality
- Maintenance (Red): Service and maintenance capabilities
- Network (Purple): Connectivity and communication features

## Reading the Diagram
- Arrows indicate dependencies (must validate source before target)
- Grouped boxes show related phases
- Each node represents a complete journey that must be validated
- Colors indicate feature category and general complexity

## Validating Journeys
1. Start from Device Bootstrap (A)
2. Follow arrows to next possible journeys
3. All incoming dependencies must be validated before starting a journey
4. Mark journey as complete in README.md after validation
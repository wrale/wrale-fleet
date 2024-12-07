# Journey Dependencies

⚠️ **IMPORTANT**: This diagram must only be updated by human operators alongside the README.md.

This diagram shows the dependencies between different user journeys that must be validated for v1.0, starting with fundamental infrastructure requirements.

```mermaid
flowchart TD
    %% Infrastructure Setup
    A[Fleet Admin: Package Build] --> B[Fleet Admin: Core Services Deploy]
    B --> C[Fleet Admin: API Services Deploy]
    C --> D[Fleet Admin: Dashboard Deploy]
    
    %% Initial Fleet Setup
    D --> E[Fleet Admin: Initial Fleet Config]
    C --> E
    E --> F[Security Admin: Initial Access Setup]

    %% First Device
    F --> G[HW Op: First Device Bootstrap]
    E --> G
    G --> H[HW Op: Basic Management]
    
    %% Basic Features with One Device
    H --> I[HW Op: Basic Metrics]
    H --> J[Fleet Admin: Basic Policy]
    H --> K[Security Admin: Basic Access Control]
    
    %% Core Features
    I --> L[HW Op: Thermal Management]
    J --> L
    
    %% Multi-Device
    H --> M[Fleet Admin: Multi-Device Enrollment]
    M --> N[Fleet Admin: Fleet-wide Policy]
    L --> N
    
    %% Advanced Features
    L --> O[HW Op: Power Management]
    N --> P[Fleet Admin: Resource Optimization]
    O --> P
    
    %% Service Management
    H --> Q[Maintenance: Basic Diagnostics]
    L --> R[Maintenance: Thermal Service]
    O --> S[Maintenance: Power Service]
    Q --> T[Maintenance: Scheduled Service]
    R --> T
    S --> T
    
    %% Network Features
    H --> U[Network Admin: Basic Connectivity]
    M --> V[Network Admin: Fleet Communication]
    V --> W[Network Admin: Performance Optimization]
    
    %% Visual grouping
    subgraph Infrastructure ["Phase 1: Infrastructure Setup"]
        A
        B
        C
        D
    end
    
    subgraph FleetInit ["Phase 2: Fleet Initialization"]
        E
        F
        G
        H
    end
    
    subgraph SingleDevice ["Phase 3: Single Device Features"]
        I
        J
        K
        L
    end
    
    subgraph MultiDevice ["Phase 4: Multi-Device Features"]
        M
        N
        O
        P
    end
    
    subgraph Maintenance ["Phase 5: Maintenance"]
        Q
        R
        S
        T
    end
    
    subgraph Network ["Phase 6: Network"]
        U
        V
        W
    end

    %% Styling
    classDef infrastructure fill:#fdd,stroke:#333,stroke-width:2px
    classDef fleetInit fill:#f9f,stroke:#333,stroke-width:2px
    classDef singleDevice fill:#bbf,stroke:#333,stroke-width:2px
    classDef multiDevice fill:#bfb,stroke:#333,stroke-width:2px
    classDef maintenance fill:#fbb,stroke:#333,stroke-width:2px
    classDef network fill:#fbf,stroke:#333,stroke-width:2px
    
    class A,B,C,D infrastructure
    class E,F,G,H fleetInit
    class I,J,K,L singleDevice
    class M,N,O,P multiDevice
    class Q,R,S,T maintenance
    class U,V,W network
```

## Node Key
- Infrastructure (Red): Foundational services and deployments
- Fleet Init (Pink): Initial fleet and device setup
- Single Device (Blue): Features validated with one device
- Multi Device (Green): Features requiring multiple devices
- Maintenance (Red): Service and maintenance capabilities
- Network (Purple): Connectivity and communication features

## Reading the Diagram
- Arrows indicate strict dependencies (must validate source before target)
- Grouped boxes show related phases
- Each node represents a complete journey that must be validated
- Colors indicate feature category and general progression

## Validation Process
1. Start with Package Build (A)
2. Follow arrows to next possible journeys
3. All incoming dependencies must be validated before starting a journey
4. Cannot skip phases - infrastructure and initialization are required
5. Mark journeys as complete in README.md after validation

## Integration Points
- Phase 1 validates core infrastructure
- Phase 2 enables first device connection
- Phase 3 establishes basic functionality
- Phase 4 scales to multiple devices
- Phase 5 adds maintenance capabilities
- Phase 6 optimizes communication
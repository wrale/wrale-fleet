# Wrale Fleet Architecture Diagrams

This document contains key architectural diagrams visualizing system components and their interactions.

## System Overview

```mermaid
graph TB
    %% Top Level Architecture
    subgraph WraleFleet["Wrale Fleet System"]
        direction TB
        
        subgraph UserLayer["User Layer"]
            dashboard["Dashboard (Next.js)"]
            api["API (Go)"]
        end
        
        subgraph FleetLayer["Fleet Layer"]
            brain["Fleet Brain"]
            edge["Fleet Edge"]
        end
        
        subgraph SyncLayer["Sync Layer"]
            syncMgr["Sync Manager"]
            resolver["Conflict Resolver"]
            store["State Store"]
        end
        
        subgraph MetalLayer["Metal Layer"]
            metald["Metal Daemon"]
            hw["Hardware Control"]
            diag["Diagnostics"]
        end
    end

    %% Layer Connections
    dashboard --> api
    api --> brain
    brain --> edge
    brain --> syncMgr
    edge --> metald
    metald --> hw

    %% Cross-layer Data Flow
    hw --> diag
    diag --> brain
    syncMgr --> edge
    resolver --> store
    store --> brain
end
```

## Metal Layer Architecture

```mermaid
graph TB
    subgraph MetalSystem["Metal Layer"]
        direction TB
        
        %% Core Components
        subgraph Core["Core (metald)"]
            server["HTTP Server"]
            policy["Policy Engine"]
            stateManager["State Manager"]
        end

        %% Hardware Control
        subgraph Hardware["Hardware Control"]
            gpio["GPIO Control"]
            power["Power Management"]
            thermal["Thermal Control"]
            security["Security Monitor"]
        end

        %% Diagnostics
        subgraph Diagnostics["Diagnostics"]
            monitor["System Monitor"]
            tests["Hardware Tests"]
            metrics["Metrics Collection"]
        end
    end

    %% External Interfaces
    Hardware --> Core
    Core --> Diagnostics
```

## Fleet Layer Architecture

```mermaid
graph TB
    subgraph FleetSystem["Fleet Layer"]
        direction TB
        
        %% Brain Components
        subgraph Brain["Fleet Brain"]
            coordinator["Coordinator"]
            orchestrator["Orchestrator"]
            scheduler["Task Scheduler"]
        end

        %% Edge Components
        subgraph Edge["Fleet Edge"]
            agent["Edge Agent"]
            metalClient["Metal Client"]
            store["Local Store"]
        end

        %% State Management
        subgraph State["State Management"]
            inventory["Device Inventory"]
            topology["Fleet Topology"]
        end
    end

    %% Integration Flows
    Brain --> Edge
    Edge --> metalClient
    inventory --> coordinator
    topology --> orchestrator
```

## Sync Layer Architecture

```mermaid
graph TB
    subgraph SyncSystem["Sync Layer"]
        direction TB
        
        %% Core Components
        subgraph Manager["Sync Manager"]
            stateMgr["State Manager"]
            configMgr["Config Manager"]
            versionMgr["Version Manager"]
        end

        %% Resolver Components
        subgraph Resolver["Conflict Resolution"]
            detector["Conflict Detector"]
            resolver["Resolver"]
            validator["State Validator"]
        end

        %% Storage
        subgraph Storage["State Storage"]
            stateDB["State DB"]
            history["Version History"]
            snapshots["State Snapshots"]
        end
    end

    %% Data Flow
    Manager --> Resolver
    Resolver --> Storage
```

## Critical Flow Diagrams

### State Synchronization Flow
```mermaid
sequenceDiagram
    participant H as Hardware
    participant M as Metal
    participant E as Edge
    participant B as Brain
    participant S as Sync
    participant U as UI

    Note over H,U: Normal Operation
    H->>M: Hardware State Change
    M->>E: Device State Update
    E->>B: Edge State Sync
    B->>S: Version State
    S-->>U: State Change Event
    
    Note over H,U: Conflict Resolution
    E->>B: Edge State A
    B->>S: Version State A
    E->>B: Edge State B (Conflict)
    B->>S: Version State B
    S-->>S: Detect Conflict
    S-->>S: Resolve Conflict
    S->>B: Resolution
    B->>E: State Update
    E->>M: Apply Changes
```

### Configuration Distribution Flow
```mermaid
sequenceDiagram
    participant U as UI
    participant A as API
    participant B as Brain
    participant S as Sync
    participant E as Edge
    participant M as Metal

    Note over U,M: Config Update
    U->>A: Update Config
    A->>B: Validate Config
    B->>S: Version Config
    S->>S: Generate Delta
    par Distribution to Edges
        S->>E: Config Delta 1
        S->>E: Config Delta 2
    end
    E->>M: Apply Config
    M-->>E: Config Applied
    E-->>S: Ack Config
    S-->>B: Distribution Complete
    B-->>A: Update Success
    A-->>U: Success Response
```

### Recovery Flow
```mermaid
sequenceDiagram
    participant H as Hardware
    participant M as Metal
    participant E as Edge
    participant B as Brain
    participant S as Sync

    Note over H,S: Recovery Process
    H->>M: Hardware Error
    M->>E: Error Event
    E->>B: Report Failure
    B->>S: Version Error State
    
    par Recovery Actions
        B->>E: Recovery Instructions
        E->>M: Recovery Commands
        M->>H: Hardware Reset
    end
    
    H-->>M: Hardware Ready
    M-->>E: Device State
    E-->>B: Recovery Complete
    B-->>S: Version Recovery State
```

### Network Partition Recovery
```mermaid
sequenceDiagram
    participant E1 as Edge 1
    participant E2 as Edge 2
    participant B as Brain
    participant S as Sync

    Note over E1,S: Partition Event
    E1->>B: State Update
    E2-xB: Update Fails
    E2-->>E2: Enter Partition Mode
    
    Note over E1,S: Recovery
    E2->>B: Reconnect
    B->>S: Request State Diff
    S->>S: Calculate Delta
    S->>B: State Resolution
    B->>E2: State Update
    E2-->>B: Ack Update
```

## Physical Safety Flows

### Temperature Safety Response
```mermaid
sequenceDiagram
    participant H as Hardware
    participant M as Metal
    participant E as Edge
    participant B as Brain
    
    Note over H,B: Temperature Event
    H->>M: Temperature Alert
    M->>M: Check Safety Bounds
    
    alt Temperature Critical
        M->>H: Emergency Shutdown
        M->>E: Critical Alert
        E->>B: Emergency Event
    else Temperature High
        M->>H: Increase Cooling
        M->>E: Warning Alert
        E->>B: Warning Event
    end
```

### Resource Management Flow
```mermaid
sequenceDiagram
    participant H as Hardware
    participant M as Metal
    participant E as Edge
    participant B as Brain
    participant S as Sync

    Note over H,S: Resource Management
    B->>S: Resource Policy Update
    S->>E: Policy Distribution
    E->>M: Resource Limits
    M->>H: Apply Constraints
    
    H-->>M: Resource Metrics
    M-->>E: Resource State
    E-->>B: Resource Usage
    B-->>S: Version Resource State
```
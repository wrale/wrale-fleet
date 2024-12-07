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

## User Layer Architecture

```mermaid
graph TB
    subgraph UserSystem["User Layer"]
        direction TB
        
        %% UI Components
        subgraph UI["Dashboard"]
            pages["Next.js Pages"]
            components["React Components"]
            services["API Services"]
        end

        %% API Components
        subgraph API["REST API"]
            handlers["Request Handlers"]
            auth["Authentication"]
            websocket["WebSocket Server"]
        end

        %% Integration
        subgraph Integration["Backend Integration"]
            fleetClient["Fleet Client"]
            syncClient["Sync Client"]
            eventBus["Event Bus"]
        end
    end

    %% Service Flow
    UI --> API
    API --> Integration
    Integration --> eventBus
```

These diagrams provide a visual representation of the key architectural components and their interactions within the Wrale Fleet system. They complement the detailed documentation in OVERVIEW.md, LAYERS.md, API.md, SECURITY.md, and DEPLOYMENT.md.
# SAGA Engine Server

SAGA (Structured Adventure Game Agent) Engine Server is a multitenant text-based adventure game server written in Go and optimized for use with tool-calling LLM agents. It is part of the SAGA Engine project, along with the [SAGA Engine Agent](https://github.com/gloomloop/saga-agent).

This repo comes with a REPL client for testing purposes.

## Quickstart

### Setup

1. Create Python environment
    ```
    python3 -m venv saga
    source saga/bin/activate
    pip install -r requirements.txt
    ```

2. Start the game server
    ```   
    go run cmd/server/main.go
    ```

3. Play the [demo game](internal/testdata/demo.yaml)
    ```
    (saga) peter@Mac server % python repl.py
    Starting Adventure Game REPL Client...
    Created session: 6752e432-1bdc-4110-afa4-d6589cc7409c
    Loaded demo game: demo puzzle

    ╔══════════════════════════════════════════════════════════════╗
    ║                    ADVENTURE GAME REPL                       ║
    ║                                                              ║
    ║  Available commands:                                         ║
    ║    observe                    - Look around the current room ║
    ║    inspect <item>             - Inspect an item or door      ║
    ║    uncover <item>             - Uncover a concealed item     ║
    ║    unlock <key/code> <target> - Unlock a door or container   ║
    ║    search <container>         - Search a container           ║
    ║    take <item>                - Take an item                 ║
    ║    inventory                  - Show your inventory          ║
    ║    heal <item>                - Use a health item            ║
    ║    go <direction/room>        - Move to another room         ║
    ║    battle <weapon>            - Battle an enemy              ║
    ║    info                       - Show session info            ║
    ║    debug                      - Show debug information       ║
    ║    quit                       - Exit the game                ║
    ║                                                              ║
    ║  Examples:                                                   ║
    ║    observe                                                   ║
    ║    inspect "tattered grey hoodie"                            ║
    ║    take "energy drink"                                       ║
    ║    go left                                                   ║
    ║    unlock "2468" "safe"                                      ║
    ╚══════════════════════════════════════════════════════════════╝
    ```

### Gameplay

- Look around

    ```
    game> observe
    {
        "engine_state": {
            "level_completion": "in_progress",
            "mode": "investigation"
        },
        "room_info": {
            "name": "waiting room",
            "description": "a dilapidated waiting room",
            "visible_items": [
                {
                    "name": "tattered grey hoodie",
                    "description": "a tattered grey hoodie",
                    "location": "middle of the floor",
                    "coneals_something": true
                },
                {
                    "name": "energy drink",
                    "description": "an unopened energy drink",
            ...
    ```

- Inspect something

    ```
    game> inspect energy drink
    {
        "engine_state": {
            "level_completion": "in_progress",
            "mode": "investigation"
        },
        "item_info": {
            "name": "energy drink",
            "description": "an unopened energy drink",
            "is_portable": true,
            "is_health_item": true,
            "details": "<text>NRG-9001: unleash your inner beast</text>"
        }
    }
    ```
- Enter the storage room
    ```
    game> go left
    {
        "engine_state": {
            "level_completion": "in_progress",
            "mode": "investigation"
        },
        "entered_room": {
            "name": "storage room",
            "description": "a dusty storage room",
            "visible_items": [
                {
                    "name": "dark green tarp",
                    "description": "a dark green tarp",
                    "location": "floor",
                    "coneals_something": true
                },
        ...
    ```
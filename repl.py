#!/usr/bin/env python3
"""
Adventure Game REPL Client

A Python REPL client for the adventure game server.
Supports all main game actions with pretty-printed JSON responses.
"""

import json
import requests
import cmd
import sys
import shlex
from typing import Optional, Dict, Any, List
from pathlib import Path

# Server configuration
SERVER_URL = "http://localhost:8080"
API_BASE = f"{SERVER_URL}/api/v1"

class GameClient:
    """Client for interacting with the adventure game server."""
    
    def __init__(self, server_url: str = SERVER_URL):
        self.server_url = server_url
        self.api_base = f"{server_url}/api/v1"
        self.session_id: Optional[str] = None
        
    def create_session(self, level_data: Dict[str, Any]) -> str:
        """Create a new game session with the provided level data."""
        url = f"{self.api_base}/sessions"
        response = requests.post(url, json={"level": level_data})
        response.raise_for_status()
        
        result = response.json()
        self.session_id = result["session_id"]
        return self.session_id
    
    def _make_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        """Make a request to the game server."""
        if not self.session_id:
            raise ValueError("No active session. Create a session first.")
            
        url = f"{self.api_base}/sessions/{self.session_id}/{endpoint}"
        
        if method.upper() == "GET":
            response = requests.get(url)
        elif method.upper() == "POST":
            response = requests.post(url, json=data or {})
        elif method.upper() == "DELETE":
            response = requests.delete(url)
        else:
            raise ValueError(f"Unsupported HTTP method: {method}")
            
        response.raise_for_status()
        return response.json()
    
    def observe(self) -> Dict[str, Any]:
        """Observe the current room."""
        return self._make_request("POST", "observe")
    
    def inspect(self, target_name: str) -> Dict[str, Any]:
        """Inspect an item or door."""
        return self._make_request("POST", "inspect", {"target_name": target_name})
    
    def uncover(self, target_name: str) -> Dict[str, Any]:
        """Uncover a concealed item."""
        return self._make_request("POST", "uncover", {"target_name": target_name})
    
    def unlock(self, key_or_code: str, target_name: str) -> Dict[str, Any]:
        """Unlock a door or container."""
        return self._make_request("POST", "unlock", {
            "key_or_code": key_or_code,
            "target_name": target_name
        })
    
    def search(self, target_name: str) -> Dict[str, Any]:
        """Search a container."""
        return self._make_request("POST", "search", {"target_name": target_name})
    
    def take(self, target_name: str) -> Dict[str, Any]:
        """Take an item."""
        return self._make_request("POST", "take", {"target_name": target_name})
    
    def inventory(self) -> Dict[str, Any]:
        """Get player inventory."""
        return self._make_request("POST", "inventory")
    
    def heal(self, health_item_name: str) -> Dict[str, Any]:
        """Use a health item."""
        return self._make_request("POST", "heal", {"health_item_name": health_item_name})
    
    def traverse(self, destination: str) -> Dict[str, Any]:
        """Move to another room."""
        return self._make_request("POST", "traverse", {"door_or_direction": destination})
    
    def battle(self, weapon_name: str) -> Dict[str, Any]:
        """Battle an enemy."""
        return self._make_request("POST", "battle", {"weapon_name": weapon_name})
    
    def get_session_info(self) -> Dict[str, Any]:
        """Get session information."""
        return self._make_request("GET", "")
    
    def debug(self) -> Dict[str, Any]:
        """Get debug information."""
        return self._make_request("GET", "debug")
    
    def delete_session(self) -> Dict[str, Any]:
        """Delete the current session."""
        if not self.session_id:
            return {"message": "No active session to delete"}
        
        url = f"{self.api_base}/sessions/{self.session_id}"
        response = requests.delete(url)
        response.raise_for_status()
        
        result = response.json()
        self.session_id = None
        return result


class GameREPL(cmd.Cmd):
    """Interactive REPL for the adventure game."""
    
    intro = """
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
║    go east                                                   ║
║    unlock "2468" "safe"                                      ║
║    unlock "iron key" "metal stairwell door"                  ║
╚══════════════════════════════════════════════════════════════╝
"""
    prompt = "game> "
    
    def __init__(self, client: GameClient):
        super().__init__()
        self.client = client
        self.print_response(self.client.observe())
    
    def parse_args(self, arg: str) -> List[str]:
        """Parse command arguments, handling quoted strings properly."""
        if not arg.strip():
            return []
        try:
            return shlex.split(arg.strip())
        except ValueError:
            # Fallback to simple split if shlex fails
            return arg.strip().split()

    def print_response(self, response: Dict[str, Any]):
        """Pretty print a JSON response."""
        print(json.dumps(response, indent=2))
        print()
    
    def do_observe(self, arg):
        """Look around the current room."""
        try:
            response = self.client.observe()
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_inspect(self, arg):
        """Inspect an item or door."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: inspect <item_name>")
            return
        
        try:
            response = self.client.inspect(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_uncover(self, arg):
        """Uncover a concealed item."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: uncover <item_name>")
            return
        
        try:
            response = self.client.uncover(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_unlock(self, arg):
        """Unlock a door or container."""
        args = self.parse_args(arg)
        if len(args) != 2:
            print("Usage: unlock <key_or_code> <target_name>")
            print("Example: unlock \"iron key\" \"metal stairwell door\"")
            return
        
        key_or_code, target_name = args
        try:
            response = self.client.unlock(key_or_code, target_name)
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_search(self, arg):
        """Search a container."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: search <container_name>")
            return
        
        try:
            response = self.client.search(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_take(self, arg):
        """Take an item."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: take <item_name>")
            return
        
        try:
            response = self.client.take(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_inventory(self, arg):
        """Show your inventory."""
        try:
            response = self.client.inventory()
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_heal(self, arg):
        """Use a health item."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: heal <health_item_name>")
            return
        
        try:
            response = self.client.heal(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_go(self, arg):
        """Move to another room."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: go <direction_or_room_name>")
            return
        
        try:
            response = self.client.traverse(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_battle(self, arg):
        """Battle an enemy."""
        args = self.parse_args(arg)
        if not args:
            print("Usage: battle <weapon_name>")
            return
        
        try:
            response = self.client.battle(" ".join(args))
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_info(self, arg):
        """Show session information."""
        try:
            response = self.client.get_session_info()
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_debug(self, arg):
        """Show debug information."""
        try:
            response = self.client.debug()
            self.print_response(response)
        except Exception as e:
            print(e.response.json().get("error"))
    
    def do_quit(self, arg):
        """Exit the game."""
        print("Thanks for playing!")
        return True
    
    def do_exit(self, arg):
        """Exit the game."""
        return self.do_quit(arg)
    
    def do_EOF(self, arg):
        """Handle Ctrl+D."""
        return self.do_quit(arg)
    
    def default(self, line):
        """Handle unknown commands."""
        print(f"Unknown command: {line}")
        print("Type 'help' for available commands.")


def main():
    """Main function to run the game REPL."""
    print("Starting Adventure Game REPL Client...")
    
    # Check if server is running
    try:
        response = requests.get(f"{SERVER_URL}/api/v1/sessions", timeout=5)
        response.raise_for_status()
    except requests.exceptions.RequestException as e:
        print(f"Error: Could not connect to server at {SERVER_URL}")
        print("Make sure the server is running on port 8080")
        sys.exit(1)
    
    # Create client and load demo game
    client = GameClient()
    
    try:
        with open("internal/testdata/demo.json", 'r') as f:
            demo_game = json.load(f)
        session_id = client.create_session(demo_game)
        print(f"Created session: {session_id}")
        print(f"Loaded demo game: {demo_game['name']}")
    except Exception as e:
        print(f"Error loading demo game: {e}")
        sys.exit(1)
    
    # Start REPL
    try:
        repl = GameREPL(client)
        repl.cmdloop()
    except KeyboardInterrupt:
        print("\nGame interrupted. Goodbye!")
    finally:
        # Clean up session
        try:
            client.delete_session()
        except:
            pass


if __name__ == "__main__":
    main()

# Connect 4

### Overview
This is a versatile Connect Four web application built with Go, supporting multiple game modes:

Local Play (Player vs Player on the same device)
Bot Play (Player vs AI)
Remote Friend Play (Invite and play with a specific friend)
Random Matchmaking (Join a global player pool)

#### Features

🎮 Multiple game modes
🤖 Adjustable AI difficulty levels
💻 Responsive web interface
🌐 Real-time multiplayer support
📱 Mobile and desktop friendly

#### Prerequisites

Go (version 1.20 or higher)
Web browser
Internet connection (for remote modes)

## Installation
Clone the Repository
bashCopygit clone https://github.com/yourusername/connect-four-webapp.git
cd connect-four-webapp
Install Dependencies
bashCopygo mod tidy
Running the Application
To start the server, simply run:
bashCopygo run main.go
By default, the application will start on http://localhost:8080
Game Modes
1. Local Play

Two players take turns on the same device
Perfect for playing with someone next to you

2. Bot Play

Play against an AI with adjustable difficulty
Difficulty levels:

Easy: Random moves
Medium: Strategic with some randomness
Hard: All Strategic

3. Remote Friend Play

Send a unique game link to your friend
Private match with direct connection

4. Random Matchmaking

Join a global pool of players
Automatically matched with an available opponent

Project Structure
```
├── go.mod               # Go module dependencies
├── main.go              # Application entry point
├── internal/
│   ├── game/            # Game logic
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── store/           # Data storage
│   └── utils/           # Utility functions
├── static/              # Static web assets
│   ├── css/             # Stylesheets
│   └── js/              # JavaScript files
└── templates/           # HTML templates
```
Configuration
Configuration can be adjusted in config.js and environment variables.
Development

### Technologies

Backend: Go
Frontend: HTML, CSS, JavaScript
Real-time Communication: WebSockets

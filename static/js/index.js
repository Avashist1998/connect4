let socket = null
const BASE_URL = "https://connect4.avashist.com"

const createGame = async (type, player1, player2, level = "") => {
    let res = await fetch(BASE_URL, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "GameType": type,
            "Player1": player1,
            "Player2": player2,
            "StartPlayer": "", // You can set this to an actual value if needed
            "Level": level
        }),
    });

    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to create the game");  // Explicitly throw an error for non-200 status codes
    }

    let data = await res.json();
    return data["match_id"];
};


const handleCreateLocalGame = () => {
    let player1 = document.getElementById("playerAName").value;
    let player2 = document.getElementById("playerBName").value;
    
    createGame("local", player1, player2)
        .then((matchID) => {
            console.log(`Player 1 ${player1}, Player 2 ${player2}, and the match ID is ${matchID}`);
            window.location.href +=  `${matchID}`; 
        })
        .catch(() => {
            alert("Something went wrong");  // Use an arrow function here to defer execution
        });
};


const handleCreateLiveGame = () => {
    createGame("live", "anonymous", "anonymous").then((matchID) => {
        console.log(`the match ID is ${matchID}`);
        window.location.href +=  `${matchID}`; 
    }).catch((e) => {
        alert("Something went wrong")
    })
}


const handleCreateBotGame = () => {
    let player1 = document.getElementById("playerANameBot").value;
    let level = "medium"
    const buttons = document.querySelectorAll(".difficulty-btn");
    const activeButton = Array.from(buttons).find(btn => btn.classList.contains("active"));
    if (activeButton) {
        level = activeButton.textContent.trim();
    }
    createGame("bot", player1, "bot", level).then((matchID) => {
        console.log(`the match ID is ${matchID}`);
        window.location.href +=  `${matchID}`; 
    }).catch((e) => {
        alert("Something went wrong")
    })
}


// Initialize WebSocket connection and set up message handling
const initializeSocket = (onOpenCallback) => {
    socket = new WebSocket("wss://connect4.avashist.com/ws/lobby");

    socket.onmessage = function (event) {
        let messageData = JSON.parse(event.data);
        console.log(messageData)
        if (messageData.type === "Join Player") {
            console.log(messageData)
            const lobby = document.getElementById("lobby")
            if (lobby) {
                lobby.addPlayer(
                    {
                        id: messageData.playerId, 
                        time: new Date(messageData.joinTime).toLocaleTimeString(),
                    })
            } else {
                console.error("Lobby element is not found")
                const lobby = document.querySelector("game-lobby");
                if (lobby) {
                    lobby.removePlayer(messageData.playerId);
                }
            }
        } else if (messageData.type == "Leave Player") {

        } else if (messageData.type == "Match Info") {
            if (messageData.matchID != undefined) {
                window.location.href += `${messageData.matchID}`
            }
        } else {

        }
    };

    socket.onerror = function (error) {
        console.error("WebSocket Error:", error);
    };
    
    socket.onclose = function () {
        console.log("WebSocket connection closed");
        // window.location.href = "https://connect4.avashist.com"
    };
    socket.onopen = function () {
        console.log("WebSocket connection opened");
        if (onOpenCallback) {
            onOpenCallback(); // Call the callback once the connection is open
        }
    };
};

const handleRandomPlay = () => {
    
    const selection = document.getElementById("gameSelection")
    if (selection) {
        selection.style.display = "none"
    }
    lobby.show()
    initializeSocket(() => {
        socket.send(JSON.stringify({"message": "Joining the lobby"}))
    })
}


window.addEventListener("beforeunload", () => {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close();
    }
});


document.addEventListener("DOMContentLoaded", () => {
    const lobby = document.getElementById("lobby")
    if (lobby) {
        lobby.hide()
    }
});

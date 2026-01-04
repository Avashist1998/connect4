let socket = null
var playerName = ""
var playerSlot = ""
var currPlayer = ""
var currSlot = "RED"
var playerType = "" // "active" or "passive"
var matchCount = 0
var player1Wins = 0
var player2Wins = 0

const getSessionId = async () => {

    const currentSessionId = await localStorage.getItem("sessionId");
    if (currentSessionId !== null) {
        // validation the session id 
        let res = await fetch(config.httpURL + "/session/" + currentSessionId, {
            method: "GET",
            headers: {"Content-Type": "application/json"},
        });
        if (res.status != 200) {
            localStorage.removeItem("sessionId");
            await getSessionId();
            return;
        }
        
        let data = await res.json();
        console.log("Valid Session ID: " + data["session_id"]);
    }
    else {
        let res = await fetch(config.httpURL + "/session", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
        });
        let data = await res.json();
        localStorage.setItem("sessionId", data["session_id"]);
        console.log("Session ID: " + data["session_id"]);
    }
}


const updatePlayerTurn = (playerHeadingID) => {
    if (playerHeadingID == "RED") {
        document.getElementById("homePlayer").style.backgroundColor = "green"
        document.getElementById("awayPlayer").style.backgroundColor = "white"
    } else {
        document.getElementById("awayPlayer").style.backgroundColor = "green"
        document.getElementById("homePlayer").style.backgroundColor = "white"
    }
}


const joinUIUpdate = (messageData) => {

    document.getElementById("homePlayer").innerHTML = messageData["player1"];
    document.getElementById("awayPlayer").innerHTML = messageData["player2"];
    document.getElementById("homePlayerColor").classList.add(messageData["player1Color"].toLowerCase() + '-circle');
    document.getElementById("awayPlayerColor").classList.add(messageData["player2Color"].toLowerCase() + '-circle');

    document.getElementById("waitingScreen").hide();
    document.getElementById("game").style.display = 'block';
    document.getElementById("rematchModal").style.display = "none"
    document.getElementById("gameOverModal").style.display = "none"
    document.querySelector(".main").classList = "main"
    document.getElementById("viewCount").innerHTML = messageData["connection_count"];
    currPlayer = messageData.currPlayer;
    
    if (currPlayer == "player1") {
        currSlot = messageData.player1Color;
    } else {
        currSlot = messageData.player2Color;
    }
    const boardElement = document.getElementById('game-board');
    boardElement.updateGrid(messageData.board);
    updatePlayerTurn(currSlot)
    // Move this to the backend
    // messageData.turn = messageData.currPlayer == document.getElementById("homePlayer").innerHTML
}


const addColumnClickListeners = () => {
    const columns = document.querySelectorAll('.col');
    columns.forEach(col => {
        col.addEventListener("click", () => {
            const col_number = Number(col.id.replace("col-", "")) 
            console.log(`Player: ${playerName} Column Clicked: ${col_number}`)
            sendMove(playerName, playerSlot, col_number)
        })
    })
}


const gameOverUpdate = (messageData) => {
    document.querySelector(".main").classList += " disabled"
    document.getElementById("gameOverModal").style.display = "block"
    if (messageData.winner === "") {
        document.getElementById("gameOverHeading").innerHTML = "Draw"
    } else {
        document.getElementById("gameOverHeading").innerHTML = `Winner: ${messageData.winner}`
    }

    if (playerType == "passive") {
        document.getElementById("rematchButton").style.display = "none"
    }
}

// Initialize WebSocket connection and set up message handling
const initializeSocket = async () => {
    const currentSessionId = await localStorage.getItem("sessionId");
    const matchId = window.location.href.split("/").pop();
    socket = new WebSocket(config.liveWSEndpoint + "?sessionId=" + currentSessionId + "&matchId=" + matchId);

    socket.onopen = function () {
        console.log("WebSocket connection opened")
        if (socket.readyState === WebSocket.OPEN) {
            console.log("Sending join message")
            socket.send(JSON.stringify({type: "join"}));
        }
    }

    socket.onmessage = function (event) {
        let messageData = JSON.parse(event.data);
        console.log(messageData)
        if (messageData.type == "game_state") {
            joinUIUpdate(messageData);
        } else if (messageData.type == "joined_game") {
            playerType = messageData.type;
        } else if (messageData.type == "chat") {
            const chatBox = document.getElementById('chat-box');
            chatBox.addMessage(messageData);
        } else if (messageData.message == "Update Game") {
            const boardElement = document.getElementById('game-board');
            boardElement.updateGrid(messageData.board);
            updatePlayerTurn(messageData.currSlot)
        } else if (messageData.type == "game_over") {
            gameOverUpdate(messageData)
        } else if (messageData.type == "rematch") {
            document.getElementById("gameOverModal").style.display = "none"
            document.getElementById("rematchModal").style.display = "block"
        } else if (messageData.type == "session_state") {
            matchCount = messageData.match_count
            player1Wins = messageData.player1_wins
            player2Wins = messageData.player2_wins
            document.getElementById("matchCount").innerHTML = matchCount
            document.getElementById("player1Wins").innerHTML = player1Wins
            document.getElementById("player2Wins").innerHTML = player2Wins
        }

    };

    socket.onerror = function (error) {
        console.error("WebSocket Error:", error);
    };
    
    socket.onclose = function () {
        console.log("WebSocket connection closed");
        // window.location.href = config.httpURL
    };
};


const sendPing = (playerName) => {
    const matchID = window.location.href.split("/").pop()
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({type: "ping"}))
    }  else {
        console.error("WebSocket is not connected");
    }
}


const sendMove = (player, slot, move) => {    
    const matchID = window.location.href.split("/").pop()
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "move", Player: player, Slot: slot, Move: move, MatchID: matchID }));
    } else {
        console.error("WebSocket is not connected");
    }
};

getSessionId().then(() => {
    const name = localStorage.getItem("name")
    if (name !== null) {
        document.getElementById("playerName").value = name;
    }
    
    // Enable/disable join button based on input
    const playerNameInput = document.getElementById("playerName");
    const joinButton = document.getElementById("joinButton");
    
    const updateJoinButtonState = () => {
        joinButton.disabled = playerNameInput.value.trim() === "";
    };
    
    playerNameInput.addEventListener("input", updateJoinButtonState);
    updateJoinButtonState(); // Set initial state
    
    // Mobile: Make match stats collapsible
    const setupMatchStatsCollapsible = () => {
        const matchStatsCard = document.getElementById("matchStatsCard");
        const matchStatsHeader = document.getElementById("matchStatsHeader");
        if (!matchStatsCard || !matchStatsHeader) return;
        
        // Store reference to avoid re-adding listeners
        if (matchStatsHeader.dataset.hasListener === 'true') return;
        
        if (window.innerWidth <= 768) {
            // Start collapsed on mobile
            matchStatsCard.classList.remove("expanded");
            matchStatsHeader.addEventListener("click", () => {
                matchStatsCard.classList.toggle("expanded");
            });
        } else {
            // Start expanded on desktop
            matchStatsCard.classList.add("expanded");
        }
        matchStatsHeader.dataset.hasListener = 'true';
    };
    
    // Setup when DOM is ready, and retry if elements aren't available yet
    const trySetupStats = () => {
        if (document.getElementById("matchStatsCard")) {
            setupMatchStatsCollapsible();
        } else {
            setTimeout(trySetupStats, 100);
        }
    };
    
    trySetupStats();
    window.addEventListener("resize", setupMatchStatsCollapsible);
    
    initializeSocket()
}).catch(() => {
    console.error("Something went wrong");
});

const onClickJoin = () => {
    const matchID = window.location.href.split("/").pop()
    playerName = document.getElementById("playerName").value;
    localStorage.setItem("name", playerName);
    document.getElementById("joinModal").style.display = "none";
    document.getElementById("waitingScreen").show();
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({type: "ready", name: playerName}))
        setInterval(() => {sendPing(playerName)}, 5000);
        addColumnClickListeners()
    } else {
        handleClickHome()
    }
}

const handleClickHome = () => {
    window.location.href =  config.httpURL
}


const handleClickRematch = () => {
    const matchID = window.location.href.split("/").pop()
    let playerName = document.getElementById("homePlayer").innerHTML;
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({type: "rematch"}));
        document.getElementById("waitingScreen").updateMessage("Waiting for rematch message")
        document.getElementById("waitingScreen").show();
        document.getElementById("gameOverModal").style.display = "none"
    } else {
        console.error("WebSocket is not connected");
    }
}


const rematchResponse = (res) => {
    const matchID = window.location.href.split("/").pop()
    let playerName = document.getElementById("homePlayer").innerHTML;
    if (socket && socket.readyState === WebSocket.OPEN) {
        if (res === true) {
            socket.send(JSON.stringify({Type: "rematch", Player: playerName, Message: "accept", MatchID: matchID}));
        } else {
            socket.send(JSON.stringify({Type: "rematch", Player: playerName, Message: "reject", MatchID: matchID}));
        }
    } else {
        console.error("WebSocket is not connected");
    }
}

document.getElementById("copyButton").addEventListener("click", () => {
    navigator.clipboard.writeText(window.location.href);
    alert("URL adding to clipboard, Share with your friends")
})


const sendMessage = (message) => {
    const matchID = window.location.href.split("/").pop()
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "chat", name: playerName, message: message}));
    } else {
        console.error("WebSocket is not connected");
    }
};

document.addEventListener("chat-message", (event) => {
    let message = event.detail.message;
    sendMessage(message);
})
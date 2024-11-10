let socket = null


const updateGrid = (board) => {
    for (let [index, val] of board.entries()) {
        cell = document.getElementById(`cell-${index}`)
        if (val == 1) {
            cell.className = 'cell player1';
        } else if (val == -1) {
            cell.className = 'cell player2';
        } else {
            cell.className = 'cell empty';
        }
    }
}


const updatePlayerTurn = (currPlayer) => {
    let playerName = document.getElementById("homePlayer").innerHTML
    console.log(currPlayer, playerName)
    if (currPlayer == playerName) {
        document.getElementById("homePlayer").style.backgroundColor = "green"
        document.getElementById("awayPlayer").style.backgroundColor = "white"
    } else {
        document.getElementById("awayPlayer").style.backgroundColor = "green"
        document.getElementById("homePlayer").style.backgroundColor = "white"
    }
}


const joinUIUpdate = (messageData) => {
    let playerName = document.getElementById("playerName").value;
    document.getElementById("homePlayer").innerHTML = playerName;
    if (playerName == messageData["player1"]) {
        document.getElementById("awayPlayer").innerHTML = messageData["player2"] 
    } else {
        document.getElementById("awayPlayer").innerHTML = messageData["player1"]
    }
    document.getElementById("loading").style.display = 'none';
    document.getElementById("game").style.display = 'block';
    document.getElementById("rematchModal").style.display = "none"
    document.querySelector(".nav-modal").style.display = "none"
    document.querySelector(".main").classList = "main"
    updateGrid(messageData.board);
    updatePlayerTurn(messageData.currPlayer);
}


const addColumnClickListeners = () => {
    const columns = document.querySelectorAll('.col');
    columns.forEach(col => {
        col.addEventListener("click", () => {
            const col_number = Number(col.id.replace("col-", ""))
            let playerName = document.getElementById("playerName").value; 
            console.log(`Player: ${playerName} Column Clicked: ${col_number}`)
            sendMove(playerName, col_number)
        })
    })
}


const gameOverUpdate = (messageData) => {
    document.querySelector(".main").classList += " disabled"
    document.querySelector(".nav-modal").style.display = "block"
    if (messageData.winner === "") {
        document.getElementById("gameOverHeading").innerHTML = "Draw"
    } else {
        document.getElementById("gameOverHeading").innerHTML = `Winner: ${messageData.winner}`
    }
}

// Initialize WebSocket connection and set up message handling
const initializeSocket = () => {
    socket = new WebSocket("wss://connect4.avashist.com/ws/live");

    socket.onmessage = function (event) {
        let messageData = JSON.parse(event.data);
        console.log(messageData)
        if (messageData.message == "Game Started") {
            joinUIUpdate(messageData);
        } else if (messageData.message == "Update Game") {
            updateGrid(messageData.board);
            updatePlayerTurn(messageData.currPlayer);
        } else if (messageData.message == "Game Over") {
            gameOverUpdate(messageData)
        } else if (messageData.message == "ReMatch Request") {
            document.querySelector(".nav-modal").style.display = "none"
            document.getElementById("rematchModal").style.display = "block"
        }
    };

    socket.onerror = function (error) {
        console.error("WebSocket Error:", error);
    };
    
    socket.onclose = function () {
        console.log("WebSocket connection closed");
        window.location.href = "https://connect4.avashist.com"
    };
};


initializeSocket()


const sendPing = (playerName) => {
    const matchID = window.location.href.split("/").pop()
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "ping", Player: playerName, MatchID: matchID}))
    }  else {
        console.error("WebSocket is not connected");
    }
}


const sendMove = (player, move) => {    
    const matchID = window.location.href.split("/").pop()
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "move", Player: player, Move: move, MatchID: matchID }));
    } else {
        console.error("WebSocket is not connected");
    }
};


const onClickJoin = () => {
    const matchID = window.location.href.split("/").pop()
    let playerName = document.getElementById("playerName").value;
    document.getElementById("joinModal").style.display = "none";
    document.getElementById("loading").style.display = "block";
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "join", Player: playerName, MatchID: matchID}))
        setInterval(() => {sendPing(playerName)}, 5000);
        addColumnClickListeners()

    }
}

const handleClickHome = () => {
    window.location.href = "https://connect4.avashist.com"
}

const createGame = async (player1, player2) => {
    let res = await fetch("https://connect4.avashist.com/match", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            "Player1": player1,
            "Player2": player2,
            "StartPlayer": "",
        }),
    });

    if (res.status != 200) {
        console.log("Some error occurred with the call");
        throw new Error("Failed to create the game");
    }

    let data = await res.json();
    return data["match_id"];
};


const handleClickRematch = () => {
    const matchID = window.location.href.split("/").pop()
    let playerName = document.getElementById("homePlayer").innerHTML;
    console.log("This was triggered ", playerName)
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({Type: "rematch", Player: playerName, Message: "request", MatchID: matchID}));
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
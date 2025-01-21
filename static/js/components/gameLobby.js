class GameLobby extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({mode: "open"})
        this._waitingQueue = [];
    }
    set waitingQueue(queue) {
        this._waitingQueue = queue
        this.render();   
    }

    get waitingQueue() {
        return this._waitingQueue;
    }

    connectedCallback() {
        this.render();
    }

    render() {
        if (!this._waitingQueue){
            this._waitingQueue = []
        }

        const tableRows = this._waitingQueue
            .map(
                (player) => `
                <tr scope="row">
                    <td>${player.id}</td>
                    <td>${player.time}</td>
                </tr>` 
            ).join("");

        this.shadowRoot.innerHTML =`
            <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css">
            <style>
                .title {
                    padding-top: 20px;
                    color: #7C444F;
                }
            </style>
            <div class="container" id="lobby">
                <h2 class="title">Lobby<h2>
                <table class="table table-striped">
                    <thead>
                    <tr>
                        <th scope="col">Player ID</th>
                        <th scope="col">Joined Time</th>
                    </tr>
                    </thead>
                    <tbody>
                        ${tableRows}
                    </tbody>
                </table>
            </div>`
    }

    addPlayer(player) {
        console.log("Adding player: ", player)
        if (!this._waitingQueue.find((p) => p.id === player.id)) {
            this._waitingQueue.push(player);
            console.log("Updated waiting queue:", this._waitingQueue);
            this.render();
        }
    }

    removePlayer(playerId) {
        // Filter out the player by ID
        this._waitingQueue = this._waitingQueue.filter((player) => player.id !== playerId);
        this.render();
    }

    show() {
        this.shadowRoot.getElementById('lobby').style.display = 'block';
    }

    hide() {
        this.shadowRoot.getElementById('lobby').style.display = 'none';
    }

}

customElements.define('game-lobby', GameLobby);
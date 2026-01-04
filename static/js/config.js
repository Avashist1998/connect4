const config = {

    SSL : false,
    BASE_URL : "127.0.0.1:9080",

    // Computed URLs
    get httpURL() {
        return this.SSL ? `https://${this.BASE_URL}` : `http://${this.BASE_URL}`;
    },

    get wsURL() {
        return this.SSL ? `wss://${this.BASE_URL}` : `ws://${this.BASE_URL}`;
    },

    get lobbyWSEndpoint() {
        return `${this.wsURL}/ws/lobby`;
    },

    get liveWSEndpoint() {
        return `${this.wsURL}/ws/live`
    }
}

Object.freeze(config)
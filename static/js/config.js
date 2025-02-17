const mode = "DEV"
const config = {

    SSL : mode == "DEV" ? false : true,
    BASE_URL : mode == "DEV" ? "192.168.1.65:9080" : "connect4.avashist.com",

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
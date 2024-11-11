class WaitingScreen extends HTMLElement {

    constructor() {
        super();
        this.attachShadow({mode: "open"})
    }

    static get observedAttributes() {
        return ['message'];
    }

    connectedCallback() {
        this.shadowRoot.innerHTML =`
            <style>
                #loadingScreen {
                    display: none;
                }

                #loading {
                    position: fixed;
                    top: 50%;
                    left: 50%;
                    transform: translate(-50%, -50%);
                    z-index: 1000;
                }

                .spinner {
                    border: 4px solid rgba(0, 0, 0, 0.1);
                    border-top: 4px solid #3498db;
                    border-radius: 50%;
                    width: 30px;
                    height: 30px;
                    animation: spin 1s linear infinite;
                }
                
                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            </style>
            <div id="loadingScreen">
                <h2>${this.message || "Waiting..."}<h2>
                <div id="loading">
                    <div class="spinner"></div>
                </div>
            </div>`
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'message') {
            this.message = newValue;
            if (this.shadowRoot) {
                this.shadowRoot.querySelector('h2').textContent = newValue;
            }
        }
    }

    updateMessage(message) {
        this.setAttribute("message", message)
    }

    show() {
        this.shadowRoot.getElementById('loadingScreen').style.display = 'block';
    }

    hide() {
        this.shadowRoot.getElementById('loadingScreen').style.display = 'none';
    }
}

customElements.define('waiting-screen', WaitingScreen);
// web/static/js/ws.js
// Alpine.js JavaScript plugin for WebSocket OOB updates.
// Automatically connects to /ws and feeds received HTML into HTMX OOB swap.
document.addEventListener('alpine:init', () => {
    Alpine.plugin((Alpine) => {
        let ws;
        let reconnectDelay = 1000;
        const maxReconnectDelay = 30000;

        function connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
            const url = `${protocol}://${window.location.host}/ws`;

            ws = new WebSocket(url);

            ws.onmessage = (event) => {
                // HTMX OOB: inject HTML fragment directly into the DOM
                // htmx.process(htmx.parseHTML(`<div>${event.data}</div>`)[0]);
                // Simplified for HTMX 2.x which handles OOB swaps natively if correctly structured
                // We wrap it in a temporary div to parse and process.
                const fragment = htmx.parseHTML(event.data);
                if (fragment && fragment.length > 0) {
                    htmx.append(document.body, event.data);
                }
            };

            ws.onclose = () => {
                console.log('WS: Connection closed. Reconnecting...', reconnectDelay);
                setTimeout(() => {
                    reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay);
                    connect();
                }, reconnectDelay);
            };

            ws.onerror = (err) => {
                console.error('WS: Connection error', err);
                ws.close();
            };

            ws.onopen = () => {
                console.log('WS: Connected');
                reconnectDelay = 1000;
            };
        }

        // Only connect if the session cookie is present
        if (document.cookie.includes('moolah_token')) {
            connect();
        }
    });
});

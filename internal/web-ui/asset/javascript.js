import * as server from './server.js'

const SERVER_HOST = 'localhost:8080'
const SERVER_URL = `http://${SERVER_HOST}`

const connection = server.connect(`ws://${SERVER_HOST}/rt`)

export const searchHistory = async (query) => {
    try {
        const qs = new URLSearchParams()
        qs.append('query', query)
        const url = new URL(`${SERVER_URL}/history-commands-list?` + qs.toString())
        const response = await fetch(url.toString())
        document.getElementById('history-commands-list').innerHTML = await response.text()
        setQueryParam('query', query)
    } catch (error) {
        console.error(error)
    }
}

export const getQueryParam = (name) => {
    return new URL(window.location.href).searchParams.get(name)
}

const setQueryParam = (name, value) => {
    const current = new URL(window.location.href)
    current.searchParams.set(name, value)
    if (!value) current.searchParams.delete(name)
    window.history.pushState({ path: current.toString() }, '', current.toString())
}

export const connectRaw = async (raw) => {
    const currentSearch = getQueryParam('query')
    const url = new URL(`${SERVER_URL}/connect-history-item`)
    const response = await fetch(url.toString(), {
        method: 'post',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ raw, query: currentSearch })
    })
    document.getElementById('history-commands-list').innerHTML = await response.text()
}

export const disconnectRaw = async (raw) => {
    const currentSearch = getQueryParam('query')
    const url = new URL(`${SERVER_URL}/disconnect-history-item`)
    const response = await fetch(url.toString(), {
        method: 'post',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ raw, query: currentSearch })
    })
    document.getElementById('history-commands-list').innerHTML = await response.text()
}

connection.addMessageListener(server.connectionHealthIndicator)
connection.addMessageListener(connectionsUpdates)

const isNullish = (x) => (x === undefined || x === null)

async function connectionsUpdates(message) {
    if (message.type === 'connections-changed') {
        await searchHistory(getQueryParam('query'))
    }
}

/**
 * @param {HTMLElement} elem
 */
const extractActionFromDomElement = (elem) => {
    if (!elem) return
    const action = elem.getAttribute('data-action')
    if (!action) return
    const actionParams = isNullish(elem.getAttribute('data-action-params')) ? undefined : elem.getAttribute('data-action-params').split(',').map(x => x.trim())
    let actionPayload = elem.getAttribute('data-action-payload')
    if (isNullish(actionPayload) && !isNullish(elem.value)) {
        actionPayload = elem.value
    }
    return {
        id: action,
        params: actionParams,
        payload: actionPayload
    }
}
document.addEventListener('click', event => {
    const action = extractActionFromDomElement(event.target)
    if (!action) return
    switch (action.id) {
        case 'history-item-raw/{id}/disconnect':
            return disconnectRaw(action.payload)
        case 'history-item-raw/{id}/connect':
            return connectRaw(action.payload)
    }
})

window.addEventListener('keypress', event => {
    const historySearch = document.getElementById('history-search')
    if (historySearch) {
        if (document.activeElement !== historySearch) {
            historySearch.value += event.key
        }
        return searchHistory(historySearch.value)
    }
})

window.addEventListener('keydown', event => {
    const historySearch = document.getElementById('history-search')
    if (historySearch && (event.key === 'Backspace' || event.code === 'Delete')) {
        if (document.activeElement !== historySearch) {
            historySearch.value = historySearch.value.slice(0, -1)
        }
        return searchHistory(historySearch.value || '')
    }
})
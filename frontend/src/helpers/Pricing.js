const GetPricing = async () => {
    const r = await fetch("/api/pricing", {
        method: 'GET',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        referrerPolicy: 'no-referrer'
    })
    return await r.json()
}

const LookupPrice = (pricing, planName) => {
    var returned = {}
    pricing.forEach(element => {
        if (element.name === planName) {
            returned = element
        }
    })
    return returned
}

export {
    GetPricing,
    LookupPrice
}
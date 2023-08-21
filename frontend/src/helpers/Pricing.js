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

export {
    GetPricing
}
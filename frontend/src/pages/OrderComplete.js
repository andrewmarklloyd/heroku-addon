import {loadStripe} from '@stripe/stripe-js';
const stripePromise = loadStripe(process.env.REACT_APP_STRIPE_PUBLIC_KEY);


const OrderComplete = () => {

    const a = () => {
            // fetch("/api/new-instance", {
    //   method: 'POST',
    //   credentials: 'same-origin',
    //   headers: {
    //     'Content-Type': 'application/json'
    //   },
    //   referrerPolicy: 'no-referrer',
    //   body: JSON.stringify({"name": location.state.name, "plan": location.state.plan})
    // })
    // .then(r => r.json())
    // .then(r => {
    //   if (r.status === 'success') {
    //     setOpen(true)
    //     setTimeout(() => {
    //       navigate("/")  
    //     }, 1000);
    //   } else {
    //     alert("failed to create nothing: " + r)
    //   }
    // })
    }
    stripePromise.then((stripe) => {
        const clientSecret = new URLSearchParams(window.location.search).get(
        'payment_intent_client_secret'
        );

        stripe.retrievePaymentIntent(clientSecret).then(({paymentIntent}) => {
            const message = document.querySelector('#message')
        
            switch (paymentIntent.status) {
            case 'succeeded':
                message.innerText = 'Success! Payment received.';
                break;
        
            case 'processing':
                message.innerText = "Payment processing. We'll update you when payment is received.";
                break;
        
            case 'requires_payment_method':
                message.innerText = 'Payment failed. Please try another payment method.';
                break;
        
            default:
                message.innerText = 'Something went wrong.';
                break;
            }
        });
    })

  return (
    <>
        <h1>Order Complete</h1>
        <div id="message"></div>
    </>
  );
}

export {
  OrderComplete
};
  
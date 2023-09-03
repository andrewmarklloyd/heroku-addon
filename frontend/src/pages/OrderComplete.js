import {loadStripe} from '@stripe/stripe-js';
const stripePromise = loadStripe(process.env.REACT_APP_STRIPE_PUBLIC_KEY);


const OrderComplete = () => {
    stripePromise.then((stripe) => {
        const clientSecret = new URLSearchParams(window.location.search).get('payment_intent_client_secret');
        if (clientSecret === "free") {
            const message = document.querySelector('#message')
            message.innerText = 'Success! Your instance is being provisioned and will be ready soon. Check the instances page for details.';
            return
        }

        stripe.retrievePaymentIntent(clientSecret).then(({paymentIntent}) => {
            const message = document.querySelector('#message')
        
            switch (paymentIntent.status) {
            case 'succeeded':
                message.innerText = 'Success! Payment received. Your instance is being provisioned and will be ready soon. Check the instances page for details.';
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
  
import {useStripe, useElements, PaymentElement} from '@stripe/react-stripe-js';

const CheckoutForm = () => {
  const stripe = useStripe();
  const elements = useElements();

  const handleSubmit = async (event) => {
    event.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    var baseURL = "http://localhost:8080"
    if (process.env.REACT_APP_BASE_URL) {
      baseURL = process.env.REACT_APP_BASE_URL
    }

    const result = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${baseURL}/order/complete`,
      },
    });

    if (result.error) {
      console.log(result.error.message);
    } else {
      // Your customer will be redirected to your `return_url`.
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <PaymentElement />
      <button disabled={!stripe}>Submit</button>
    </form>
  )
};

export default CheckoutForm;

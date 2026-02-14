// Clover SDK integration
let clover;
let cloverConfig = {};

document.addEventListener("DOMContentLoaded", onDOMContentLoaded);

async function onDOMContentLoaded() {
  // Load config
  try {
    const response = await fetch("/api/config");
    cloverConfig = await response.json();
    console.log("Config loaded, publicKey:", cloverConfig.publicKey ? "present" : "missing");
    console.log("Config loaded, merchantId:", cloverConfig.merchantId ? "present" : "missing");
  } catch (error) {
    console.error("Failed to load config:", error);
    return;
  }

  if (!cloverConfig.publicKey) {
    console.error("Clover public key not configured");
    showError("Payment system not configured");
    return;
  }

  // Initialize Clover
  try {
    console.log("Initializing Clover SDK...");
    clover = new Clover(cloverConfig.publicKey, {
      merchantId: cloverConfig.merchantId,
    });
    const elements = clover.elements();

    // Create card input fields
    const cardNumber = elements.create("CARD_NUMBER");
    const cardDate = elements.create("CARD_DATE");
    const cardCvv = elements.create("CARD_CVV");
    const cardPostalCode = elements.create("CARD_POSTAL_CODE");

    // Mount to DOM
    cardNumber.mount("#card-number");
    cardDate.mount("#card-date");
    cardCvv.mount("#card-cvv");
    cardPostalCode.mount("#card-postal-code");

    console.log("Clover SDK initialized successfully");
  } catch (error) {
    console.error("Failed to initialize Clover:", error);
    showError("Failed to load payment form");
    return;
  }

  // Setup other functionality
  setupProductSelection();
  setupFormSubmission();
}

function setupProductSelection() {
  const productInputs = document.querySelectorAll('input[name="product"]');
  const totalAmount = document.getElementById("total-amount");

  productInputs.forEach((input) => {
    input.addEventListener("change", () => {
      const price = parseInt(input.dataset.price, 10);
      totalAmount.textContent = formatPrice(price);
    });
  });
}

function setupFormSubmission() {
  const form = document.getElementById("payment-form");
  const submitBtn = document.getElementById("submit-btn");
  const btnText = document.getElementById("btn-text");
  const btnLoading = document.getElementById("btn-loading");

  form.addEventListener("submit", function(event) {
    event.preventDefault();
    console.log("Form submitted");
    clearError();

    if (!clover) {
      showError("Payment system not initialized");
      return;
    }

    // Get form data
    const formData = new FormData(form);
    const selectedProduct = form.querySelector('input[name="product"]:checked');
    const amount = parseInt(selectedProduct.dataset.price, 10);

    // Disable button
    submitBtn.disabled = true;
    btnText.classList.add("hidden");
    btnLoading.classList.remove("hidden");

    console.log("Creating token...");
    clover.createToken()
      .then(function(result) {
        console.log("Token result:", result);

        if (result.errors) {
          Object.values(result.errors).forEach(function(error) {
            showError(error);
          });
          resetButton();
          return;
        }

        // Send to backend
        return fetch("/api/checkout", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            token: result.token,
            amount: amount,
            product: formData.get("product"),
            customer: {
              name: formData.get("name"),
              email: formData.get("email"),
            },
          }),
        });
      })
      .then(function(response) {
        if (!response) return;
        return response.json().then(function(data) {
          if (!response.ok) {
            throw new Error(data.error || "Payment failed");
          }
          showSuccess();
        });
      })
      .catch(function(error) {
        console.error("Payment error:", error);
        showError(error.message || "Payment failed. Please try again.");
        resetButton();
      });

    function resetButton() {
      submitBtn.disabled = false;
      btnText.classList.remove("hidden");
      btnLoading.classList.add("hidden");
    }
  });
}

function formatPrice(cents) {
  return "$" + (cents / 100).toFixed(2);
}

function showError(message) {
  const errorEl = document.getElementById("card-response");
  errorEl.textContent = message;
  errorEl.style.color = "red";
}

function clearError() {
  const errorEl = document.getElementById("card-response");
  errorEl.textContent = "";
}

function showSuccess() {
  document.getElementById("payment-form").classList.add("hidden");
  document.getElementById("success-message").classList.remove("hidden");
}

// Cart Drawer
function openCart(prefetchedData) {
  var drawer = document.getElementById("cart-drawer");
  var backdrop = document.getElementById("cart-backdrop");
  backdrop.classList.remove("hidden");
  // Force reflow before removing translate so transition fires
  drawer.offsetHeight;
  drawer.classList.remove("translate-x-full");
  document.body.style.overflow = "hidden";

  if (prefetchedData) {
    renderCart(prefetchedData);
    return;
  }

  fetch("/api/cart", { headers: { Accept: "application/json" } })
    .then(function (res) { return res.json(); })
    .then(function (data) { renderCart(data); })
    .catch(function () {
      document.getElementById("cart-items").innerHTML =
        '<p class="text-destructive text-center py-8">Failed to load cart.</p>';
    });
}

function closeCart() {
  var drawer = document.getElementById("cart-drawer");
  var backdrop = document.getElementById("cart-backdrop");
  drawer.classList.add("translate-x-full");
  backdrop.classList.add("hidden");
  document.body.style.overflow = "";
}

function updateNavCartCount(count) {
  var countEl = document.getElementById("nav-cart-count");
  var mobileCountEl = document.getElementById("nav-cart-mobile-count");
  var text = count > 0 ? " (" + count + ")" : "";
  if (countEl) countEl.textContent = text;
  if (mobileCountEl) mobileCountEl.textContent = text;
}

function renderCart(data) {
  var container = document.getElementById("cart-items");
  var footer = document.getElementById("cart-footer");

  updateNavCartCount(data.cartCount);

  if (!data.items || data.items.length === 0) {
    container.innerHTML =
      '<div class="text-center py-8">' +
        '<p class="text-muted-foreground mb-4">Your cart is empty.</p>' +
        '<a href="/shop" class="text-primary hover:underline">Browse the shop</a>' +
      '</div>';
    footer.classList.add("hidden");
    return;
  }

  var html = '<div class="space-y-4">';
  for (var i = 0; i < data.items.length; i++) {
    var item = data.items[i];
    var subtotal = (item.subtotalCents / 100).toFixed(2);
    var unitPrice = (item.priceCents / 100).toFixed(2);
    html +=
      '<div class="flex items-start justify-between gap-3 pb-4 border-b border-foreground/10">' +
        '<div class="grow">' +
          '<p class="font-serif font-bold text-sm">' + escapeHtml(item.name) + '</p>' +
          '<p class="text-xs text-muted-foreground">$' + unitPrice + ' each</p>' +
          (item.stockLimit >= 0 ? '<p class="text-xs text-destructive">' + item.stockLimit + ' available</p>' : '') +
        '</div>' +
        '<div class="flex items-center gap-1">' +
          '<button data-cart-qty="' + escapeAttr(item.productId) + '" data-qty="' + (item.quantity - 1) + '" class="w-7 h-7 flex items-center justify-center border border-border rounded text-sm hover:border-primary cursor-pointer">−</button>' +
          '<span class="w-8 text-center text-sm">' + item.quantity + '</span>' +
          '<button data-cart-qty="' + escapeAttr(item.productId) + '" data-qty="' + (item.quantity + 1) + '"' + (item.stockLimit >= 0 && item.quantity >= item.stockLimit ? ' disabled' : '') + ' class="w-7 h-7 flex items-center justify-center border border-border rounded text-sm hover:border-primary cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed">+</button>' +
        '</div>' +
        '<div class="text-right min-w-[60px]">' +
          '<p class="text-sm font-bold">$' + subtotal + '</p>' +
          '<button data-cart-remove="' + escapeAttr(item.productId) + '" class="text-xs text-destructive hover:underline mt-1 cursor-pointer">Remove</button>' +
        '</div>' +
      '</div>';
  }
  html += '</div>';

  container.innerHTML = html;

  var totalDollars = (data.totalCents / 100).toFixed(2);
  document.getElementById("cart-total").textContent = "$" + totalDollars;
  footer.classList.remove("hidden");
}

function escapeHtml(text) {
  var div = document.createElement("div");
  div.appendChild(document.createTextNode(text));
  return div.innerHTML;
}

// Escape a string for safe use inside a JS single-quoted string in an onclick attribute.
function escapeAttr(s) {
  return String(s).replace(/\\/g, "\\\\").replace(/'/g, "\\'").replace(/</g, "\\x3c");
}

function cartMutate(url, body) {
  var formData = new FormData();
  for (var key in body) {
    formData.append(key, body[key]);
  }
  return fetch(url, {
    method: "POST",
    headers: { Accept: "application/json" },
    body: formData,
  })
    .then(function (res) { return res.json(); })
    .then(function (data) {
      renderCart(data);
      // Keep checkout page in sync when cart is modified from the drawer.
      if (document.getElementById("order-items")) {
        if (!data.items || data.items.length === 0) {
          window.location.reload();
        } else {
          renderOrderSummary(data);
        }
      }
      return data;
    });
}

function cartUpdateQty(productId, quantity) {
  cartMutate("/cart/update", { product_id: productId, quantity: quantity });
}

function cartRemove(productId) {
  cartMutate("/cart/remove", { product_id: productId });
}

// Add-to-cart: intercept forms, open drawer on success
document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll('form[action="/cart/add"]').forEach(function (form) {
    form.addEventListener("submit", function (e) {
      e.preventDefault();
      var btn = form.querySelector('button[type="submit"]');
      var origText = btn.textContent;
      btn.disabled = true;
      btn.textContent = "Added!";

      fetch("/cart/add", {
        method: "POST",
        headers: { Accept: "application/json" },
        body: new FormData(form),
      })
        .then(function (res) { return res.json(); })
        .then(function (data) {
          openCart(data);
          setTimeout(function () {
            btn.disabled = false;
            btn.textContent = origText;
          }, 1000);
        })
        .catch(function () {
          btn.disabled = false;
          btn.textContent = origText;
        });
    });
  });
});

// Close drawer on Escape key
document.addEventListener("keydown", function (e) {
  if (e.key === "Escape") closeCart();
});

// Bind cart open/close and event delegation (no inline onclick)
document.addEventListener("DOMContentLoaded", function () {
  var navCart = document.getElementById("nav-cart");
  var navCartMobile = document.getElementById("nav-cart-mobile");
  var cartBackdrop = document.getElementById("cart-backdrop");
  var cartClose = document.getElementById("cart-close");
  var cartDrawer = document.getElementById("cart-drawer");

  if (navCart) navCart.addEventListener("click", function () { openCart(); });
  if (navCartMobile) navCartMobile.addEventListener("click", function () { openCart(); });
  if (cartBackdrop) cartBackdrop.addEventListener("click", function () { closeCart(); });
  if (cartClose) cartClose.addEventListener("click", function () { closeCart(); });

  // Cart drawer: qty/remove buttons
  if (cartDrawer) cartDrawer.addEventListener("click", function (e) {
    var qtyBtn = e.target.closest("[data-cart-qty]");
    if (qtyBtn) {
      cartUpdateQty(qtyBtn.getAttribute("data-cart-qty"), parseInt(qtyBtn.getAttribute("data-qty"), 10));
      return;
    }
    var rmBtn = e.target.closest("[data-cart-remove]");
    if (rmBtn) {
      cartRemove(rmBtn.getAttribute("data-cart-remove"));
    }
  });

  // Checkout page: qty/remove buttons
  var orderSummary = document.getElementById("order-summary");
  if (orderSummary) orderSummary.addEventListener("click", function (e) {
    var qtyBtn = e.target.closest("[data-checkout-qty]");
    if (qtyBtn) {
      checkoutUpdateQty(qtyBtn.getAttribute("data-checkout-qty"), parseInt(qtyBtn.getAttribute("data-qty"), 10));
      return;
    }
    var rmBtn = e.target.closest("[data-checkout-remove]");
    if (rmBtn) {
      checkoutRemove(rmBtn.getAttribute("data-checkout-remove"));
    }
  });
});

// Checkout page cart management
function checkoutMutate(url, body) {
  var formData = new FormData();
  for (var key in body) {
    formData.append(key, body[key]);
  }
  return fetch(url, {
    method: "POST",
    headers: { Accept: "application/json" },
    body: formData,
  })
    .then(function (res) {
      if (!res.ok) {
        return res.json().then(function (data) {
          console.error("Cart error:", data.error);
          throw new Error(data.error);
        });
      }
      return res.json();
    })
    .then(function (data) {
      updateNavCartCount(data.cartCount);
      if (!data.items || data.items.length === 0) {
        window.location.reload();
        return;
      }
      renderOrderSummary(data);
    })
    .catch(function (err) {
      console.error("Cart update failed:", err);
    });
}

function checkoutUpdateQty(productId, quantity) {
  checkoutMutate("/cart/update", { product_id: productId, quantity: quantity });
}

function checkoutRemove(productId) {
  checkoutMutate("/cart/remove", { product_id: productId });
}

function renderOrderSummary(data) {
  var container = document.getElementById("order-items");
  var amountInput = document.getElementById("cart-amount");
  var totalAmountEl = document.getElementById("total-amount");

  var html = "";
  for (var i = 0; i < data.items.length; i++) {
    var item = data.items[i];
    var subtotal = (item.subtotalCents / 100).toFixed(2);
    var unitPrice = (item.priceCents / 100).toFixed(2);
    html +=
      '<div class="flex items-center justify-between py-3 border-b border-foreground/10 text-sm">' +
        '<div class="grow">' +
          '<p class="font-serif font-bold">' + escapeHtml(item.name) + '</p>' +
          '<p class="text-xs text-muted-foreground">$' + unitPrice + ' each</p>' +
          (item.stockLimit >= 0 ? '<p class="text-xs text-destructive">' + item.stockLimit + ' available</p>' : '') +
        '</div>' +
        '<div class="flex items-center gap-1">' +
          '<button data-checkout-qty="' + escapeAttr(item.productId) + '" data-qty="' + (item.quantity - 1) + '" class="w-7 h-7 flex items-center justify-center border border-border rounded text-sm hover:border-primary cursor-pointer">\u2212</button>' +
          '<span class="w-8 text-center text-sm">' + item.quantity + '</span>' +
          '<button data-checkout-qty="' + escapeAttr(item.productId) + '" data-qty="' + (item.quantity + 1) + '"' + (item.stockLimit >= 0 && item.quantity >= item.stockLimit ? ' disabled' : '') + ' class="w-7 h-7 flex items-center justify-center border border-border rounded text-sm hover:border-primary cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed">+</button>' +
        '</div>' +
        '<div class="text-right min-w-[60px] ml-3">' +
          '<p class="font-bold">$' + subtotal + '</p>' +
          '<button data-checkout-remove="' + escapeAttr(item.productId) + '" class="text-xs text-destructive hover:underline mt-1 cursor-pointer">Remove</button>' +
        '</div>' +
      '</div>';
  }
  container.innerHTML = html;

  var summaryTotals = document.getElementById("order-summary-totals");
  var feeCents = summaryTotals ? parseInt(summaryTotals.getAttribute("data-fee-cents"), 10) || 0 : 0;
  var subtotalDollars = (data.totalCents / 100).toFixed(2);
  var grandTotalCents = data.totalCents + feeCents;
  var grandTotalDollars = (grandTotalCents / 100).toFixed(2);

  var subtotalEl = document.getElementById("order-subtotal");
  var grandTotalEl = document.getElementById("order-summary-grand-total");
  if (subtotalEl) subtotalEl.textContent = "$" + subtotalDollars;
  if (grandTotalEl) grandTotalEl.textContent = "$" + grandTotalDollars;
  if (amountInput) amountInput.value = grandTotalCents;
  if (totalAmountEl) totalAmountEl.textContent = "$" + grandTotalDollars;
}

// Clover SDK integration for checkout page
let clover;
let cloverConfig = {};

document.addEventListener("DOMContentLoaded", onDOMContentLoaded);

async function onDOMContentLoaded() {
  // Only init on checkout page (where payment form exists)
  if (!document.getElementById("payment-form")) return;

  // Load config
  try {
    const response = await fetch("/api/config");
    cloverConfig = await response.json();
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
    clover = new Clover(cloverConfig.publicKey, {
      merchantId: cloverConfig.merchantId,
    });
    const elements = clover.elements();

    // Resolve site colors to rgb() for the Clover iframe (oklch vars won't work inside iframes)
    function resolveColor(cssVar) {
      var el = document.createElement("div");
      el.style.color = "var(" + cssVar + ")";
      el.style.display = "none";
      document.body.appendChild(el);
      var rgb = getComputedStyle(el).color;
      document.body.removeChild(el);
      return rgb;
    }
    var bg = resolveColor("--background");
    var fg = resolveColor("--foreground");
    var muted = resolveColor("--muted-foreground");
    var fieldStyles = {
      body: {
        fontFamily: '"Noto Sans", "Nunito", sans-serif',
        fontSize: "16px",
        color: fg,
        backgroundColor: bg,
        margin: "0",
        padding: "0",
      },
      input: {
        fontSize: "16px",
        color: fg,
        backgroundColor: "transparent",
        fontFamily: '"Noto Sans", "Nunito", sans-serif',
        border: "none",
        outline: "none",
      },
      "input::placeholder": {
        color: muted,
      },
    };

    const cardNumber = elements.create("CARD_NUMBER", fieldStyles);
    const cardDate = elements.create("CARD_DATE", fieldStyles);
    const cardCvv = elements.create("CARD_CVV", fieldStyles);
    const cardPostalCode = elements.create("CARD_POSTAL_CODE", fieldStyles);

    cardNumber.mount("#card-number");
    cardDate.mount("#card-date");
    cardCvv.mount("#card-cvv");
    cardPostalCode.mount("#card-postal-code");
  } catch (error) {
    console.error("Failed to initialize Clover:", error);
    showError("Failed to load payment form");
    return;
  }

  setupFormSubmission();
}

function setupFormSubmission() {
  const form = document.getElementById("payment-form");
  const submitBtn = document.getElementById("submit-btn");
  const btnText = document.getElementById("btn-text");
  const btnLoading = document.getElementById("btn-loading");

  form.addEventListener("submit", function (event) {
    event.preventDefault();
    clearError();

    if (!clover) {
      showError("Payment system not initialized");
      return;
    }

    // Amount comes from server-side cart total (in cents)
    const amountEl = document.getElementById("cart-amount");
    const amount = parseInt(amountEl.value, 10);
    if (!amount || amount <= 0) {
      showError("Invalid cart amount");
      return;
    }

    const formData = new FormData(form);

    submitBtn.disabled = true;
    btnText.classList.add("hidden");
    btnLoading.classList.remove("hidden");

    clover
      .createToken()
      .then(function (result) {
        if (result.errors) {
          Object.values(result.errors).forEach(function (error) {
            showError(error);
          });
          resetButton();
          return;
        }

        return fetch("/api/checkout", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            token: result.token,
            amount: amount,
            product: "cart",
            customer: {
              firstName: formData.get("first_name"),
              lastName: formData.get("last_name"),
              email: formData.get("email"),
              phone: formData.get("phone"),
              note: formData.get("note"),
            },
          }),
        });
      })
      .then(function (response) {
        if (!response) return;
        return response.json().then(function (data) {
          if (!response.ok) {
            throw new Error(data.error || "Payment failed");
          }
          showSuccess();
        });
      })
      .catch(function (error) {
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

function showError(message) {
  const errorEl = document.getElementById("card-response");
  errorEl.textContent = message;
  errorEl.style.color = getComputedStyle(document.documentElement).getPropertyValue("--destructive-foreground").trim() || "red";
}

function clearError() {
  const errorEl = document.getElementById("card-response");
  errorEl.textContent = "";
}

function showSuccess() {
  document.getElementById("payment-form").classList.add("hidden");
  document.getElementById("order-summary").classList.add("hidden");
  document.getElementById("success-message").classList.remove("hidden");
  updateNavCartCount(0);
}

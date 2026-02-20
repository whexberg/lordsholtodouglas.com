---
title: "Payment Successful"
description: "Your transaction has been completed"
---

Thank you for your payment! Your transaction has been processed successfully.

You will be redirected to the home page shortly. If you are not redirected automatically, please click the link below.

<div class="flex justify-center mt-8">
    <a 
        href="/"
        class="font-display text-sm uppercase tracking-widest text-primary hover:text-primary/80 border-b-2 border-primary/0 hover:border-primary transition-all"
    >
        Return Home →
    </a>
</div>

<script>
    setTimeout(function() {
        window.location.href = "/";
    }, 5000);
</script>

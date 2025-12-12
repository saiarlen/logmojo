document.addEventListener("DOMContentLoaded", function () {
  // Vanilla JS Slider
  const slides = document.querySelectorAll(".slide");
  const dots = document.querySelectorAll(".dot");
  let currentSlide = 0;

  function showSlide(index) {
    slides.forEach((slide, i) => {
      slide.classList.toggle("active", i === index);
    });
    dots.forEach((dot, i) => {
      dot.classList.toggle("active", i === index);
    });
  }

  function nextSlide() {
    currentSlide = (currentSlide + 1) % slides.length;
    showSlide(currentSlide);
  }

  // Auto-advance slides
  if (slides.length > 0) {
    setInterval(nextSlide, 4000);
  }

  // Dot navigation
  dots.forEach((dot, index) => {
    dot.addEventListener("click", () => {
      currentSlide = index;
      showSlide(currentSlide);
    });
  });

  // Copy functionality
  document.querySelectorAll(".copy-btn").forEach((btn) => {
    btn.addEventListener("click", function () {
      const text = this.getAttribute("data-copy");
      navigator.clipboard.writeText(text).then(() => {
        const icon = this.querySelector("i");
        icon.className = "fas fa-check";
        setTimeout(() => {
          icon.className = "fas fa-copy";
        }, 2000);
      });
    });
  });

  // Mobile menu toggle
  const mobileToggle = document.querySelector(".mobile-menu-toggle");
  const navLinks = document.querySelector(".nav-links");

  if (mobileToggle && navLinks) {
    mobileToggle.addEventListener("click", () => {
      navLinks.classList.toggle("active");
    });
  }

  // Smooth scrolling for anchor links
  document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
    anchor.addEventListener("click", function (e) {
      e.preventDefault();
      const target = document.querySelector(this.getAttribute("href"));
      if (target) {
        target.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      }
    });
  });
});
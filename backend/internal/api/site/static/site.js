(function () {
  var root = document.documentElement;
  if (!('IntersectionObserver' in window)) return;
  root.classList.add('js');
  var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  function ready() {
    // Staggered entrance: items in the same section enter 70ms apart.
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (!e.isIntersecting) return;
        var el = e.target;
        var sec = el.closest('section, footer') || document.body;
        var group = Array.prototype.filter.call(sec.querySelectorAll('.rise:not(.in)'), function (x) {
          var r = x.getBoundingClientRect();
          return r.top < window.innerHeight && r.bottom > 0;
        });
        group.forEach(function (x, i) {
          x.style.transitionDelay = (i * 70) + 'ms';
          x.classList.add('in');
          io.unobserve(x);
        });
        if (!el.classList.contains('in')) { el.classList.add('in'); io.unobserve(el); }
      });
    }, { rootMargin: '0px 0px -8% 0px', threshold: 0.05 });
    document.querySelectorAll('.rise').forEach(function (el) { io.observe(el); });

    // Hero: the sun rises, then the phones arrive.
    var heroBits = document.querySelectorAll('.sun, .phone-hero, .phone-back');
    requestAnimationFrame(function () {
      requestAnimationFrame(function () { heroBits.forEach(function (el) { el.classList.add('in'); }); });
    });

    // Scroll-linked: the sun sinks slowly as the hero leaves.
    var sun = document.querySelector('.sun');
    if (sun && !reduce) {
      var ticking = false;
      window.addEventListener('scroll', function () {
        if (ticking) return;
        ticking = true;
        requestAnimationFrame(function () {
          var y = Math.min(window.scrollY, 900) * 0.18;
          sun.style.setProperty('--sun-y', y.toFixed(1) + 'px');
          ticking = false;
        });
      }, { passive: true });
    }

    // Safety: never leave content hidden.
    setTimeout(function () {
      document.querySelectorAll('.rise:not(.in)').forEach(function (el) {
        var r = el.getBoundingClientRect();
        if (r.top < window.innerHeight) el.classList.add('in');
      });
    }, 2500);
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', ready);
  else ready();
})();

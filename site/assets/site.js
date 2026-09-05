(() => {
  const docLayout = document.querySelector(".doc-layout");
  const article = document.querySelector(".prose");
  const header = document.querySelector("[data-site-header]");

  if (docLayout) {
    document.body.classList.add("doc-page");
    initGuideNav();
  }

  initHeaderGuidesLink();

  if (header) {
    const onHeaderScroll = () => {
      header.classList.toggle("is-stuck", window.scrollY > 12);
    };
    onHeaderScroll();
    window.addEventListener("scroll", onHeaderScroll, { passive: true });
  }

  if (article && docLayout) {
    initReadingProgress(article, docLayout);
    initScrollSpy(article);
  }

  const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reduce) {
    document.querySelectorAll(".reveal").forEach((el) => el.classList.add("is-visible"));
    return;
  }

  const io = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          entry.target.classList.add("is-visible");
          io.unobserve(entry.target);
        }
      }
    },
    { threshold: 0.16 },
  );

  document.querySelectorAll(".reveal").forEach((el) => io.observe(el));
})();

function initReadingProgress(article, docLayout) {
  const rail = document.createElement("aside");
  rail.className = "reading-rail";
  rail.setAttribute("aria-label", "Reading progress");
  rail.innerHTML = `
    <span class="reading-rail__label">Reading</span>
    <span class="reading-rail__pct">0%</span>
    <div class="reading-rail__track">
      <span
        class="reading-rail__fill"
        role="progressbar"
        aria-valuemin="0"
        aria-valuemax="100"
        aria-valuenow="0"
        aria-label="Reading progress"
      ></span>
    </div>
  `;
  docLayout.appendChild(rail);

  const fill = rail.querySelector(".reading-rail__fill");
  const pctEl = rail.querySelector(".reading-rail__pct");

  const update = () => {
    const rect = article.getBoundingClientRect();
    const articleTop = window.scrollY + rect.top;
    const scrollable = article.offsetHeight - window.innerHeight;
    const scrolled = window.scrollY - articleTop;
    const pct =
      scrollable <= 0 ? 100 : Math.min(100, Math.max(0, (scrolled / scrollable) * 100));
    const rounded = Math.round(pct);

    fill.style.height = `${pct}%`;
    fill.setAttribute("aria-valuenow", String(rounded));
    pctEl.textContent = `${rounded}%`;
  };

  update();
  window.addEventListener("scroll", update, { passive: true });
  window.addEventListener("resize", update, { passive: true });
}

function initScrollSpy(article) {
  const links = [...document.querySelectorAll(".side-nav a[href^='#']")];
  if (!links.length) {
    return;
  }

  const sections = links
    .map((link) => {
      const id = link.getAttribute("href").slice(1);
      const el = document.getElementById(id);
      return el ? { id, el, link, item: link.closest("li") } : null;
    })
    .filter(Boolean);

  if (!sections.length) {
    return;
  }

  let activeId = sections[0].id;

  const setActive = (id) => {
    if (id === activeId) {
      return;
    }
    activeId = id;
    for (const section of sections) {
      const on = section.id === id;
      section.item?.classList.toggle("is-active", on);
      if (on) {
        section.link.setAttribute("aria-current", "true");
      } else {
        section.link.removeAttribute("aria-current");
      }
    }
  };

  const pickActive = () => {
    const marker =
      window.scrollY +
      parseFloat(getComputedStyle(document.documentElement).getPropertyValue("--header-offset")) +
      24;
    let current = sections[0].id;

    for (const section of sections) {
      const top = section.el.getBoundingClientRect().top + window.scrollY;
      if (top <= marker) {
        current = section.id;
      }
    }

    setActive(current);
  };

  pickActive();
  window.addEventListener("scroll", pickActive, { passive: true });
  window.addEventListener("resize", pickActive, { passive: true });
}

const GUIDE_PAGES = [
  { href: "/getting-started.html", label: "Getting started" },
  { href: "/ship.html", label: "Ship workflow" },
  { href: "/commands.html", label: "Commands" },
  { href: "/providers.html", label: "Providers" },
  { href: "/auth.html", label: "Auth" },
  { href: "/configure-and-deploy.html", label: "Configure & deploy" },
  { href: "/secrets.html", label: "Secrets & status" },
  { href: "/state-and-runs.html", label: "State & runs" },
  { href: "/troubleshooting.html", label: "Troubleshooting" },
];

function initHeaderGuidesLink() {
  const nav = document.querySelector(".site-header .nav");
  if (!nav || nav.querySelector("[data-guides-link]")) {
    return;
  }

  const link = document.createElement("a");
  link.href = "/guides.html";
  link.textContent = "Guides";
  link.dataset.guidesLink = "";
  const github = nav.querySelector('a[href*="github.com"]');
  if (github) {
    nav.insertBefore(link, github);
  } else {
    nav.appendChild(link);
  }

  const path = normalizePath(location.pathname);
  if (path === "/guides.html") {
    link.setAttribute("aria-current", "page");
  }
}

function initGuideNav() {
  const sideNav = document.querySelector(".side-nav");
  if (!sideNav || sideNav.querySelector(".side-nav__guides")) {
    return;
  }

  const path = normalizePath(location.pathname);
  const block = document.createElement("div");
  block.className = "side-nav__guides";
  block.innerHTML = "<h2>All guides</h2>";

  const list = document.createElement("ul");
  for (const page of GUIDE_PAGES) {
    const li = document.createElement("li");
    const a = document.createElement("a");
    a.href = page.href;
    a.textContent = page.label;
    if (page.href === path) {
      a.setAttribute("aria-current", "page");
    }
    li.appendChild(a);
    list.appendChild(li);
  }

  block.appendChild(list);
  sideNav.appendChild(block);
}

function normalizePath(pathname) {
  if (!pathname || pathname === "/") {
    return "/index.html";
  }
  if (pathname.endsWith("/")) {
    return `${pathname.slice(0, -1)}.html`;
  }
  if (!pathname.includes(".") && !pathname.endsWith(".html")) {
    return `${pathname}.html`;
  }
  return pathname;
}

import React, { useEffect, useRef, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

interface LandingPageProps {
  onEnterWorkspace?: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onEnterWorkspace }) => {
  const navigate = useNavigate();
  const progressBarRef = useRef<HTMLDivElement>(null);
  const heroVisualRef = useRef<HTMLDivElement>(null);
  const [motionKey, setMotionKey] = useState<number>(0);

  const handleEnterWorkspace = useCallback(() => {
    if (onEnterWorkspace) {
      onEnterWorkspace();
    } else {
      navigate('/workspace');
    }
  }, [navigate, onEnterWorkspace]);

  // Scroll Progress Tracking
  useEffect(() => {
    const handleScroll = () => {
      if (!progressBarRef.current) return;
      const totalScroll = document.documentElement.scrollHeight - window.innerHeight;
      const progress = totalScroll > 0 ? window.scrollY / totalScroll : 0;
      progressBarRef.current.style.transform = `scaleX(${progress})`;
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // Replay Map Motion handler
  const handleReplayMotion = () => {
    setMotionKey((prev) => prev + 1);
  };

  return (
    <div className="mjengo-app">
      {/* Scroll Progress Bar */}
      <div
        ref={progressBarRef}
        className="motion-progress"
        aria-hidden="true"
      />

      <div className="landing-shell">
        {/* Simple Accessible Topbar */}
        <header className="landing-topbar" role="banner">
          <a
            href="#home"
            className="brand-link"
            aria-label="Mjengo Fiti home"
          >
            <span className="brandmark-box" aria-hidden="true">
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M4 20V8l8-4 8 4v12" />
                <path d="M8 20v-6h8v6" />
                <circle cx="8" cy="9" r="0.8" fill="currentColor" />
                <circle cx="16" cy="9" r="0.8" fill="currentColor" />
              </svg>
            </span>
            <span>
              mjengo <em>fiti</em>
            </span>
          </a>

          <nav className="landing-navlinks" aria-label="Main navigation">
            <a href="#map-preview">Map experience</a>
            <a href="#workflow">Workflow</a>
          </nav>

          <div className="landing-rightnav">
            <button
              type="button"
              className="mjengo-btn-ghost"
              onClick={handleEnterWorkspace}
            >
              Sign in
            </button>
            <button
              type="button"
              className="mjengo-btn"
              onClick={handleEnterWorkspace}
            >
              Open workspace
            </button>
          </div>
        </header>

        {/* Main Content Landmark */}
        <main id="home">
          {/* Hero Section */}
          <section className="landing-hero" aria-labelledby="hero-title">
            <div>
              <span className="hero-eyebrow">For county inspection teams</span>
              <h1 id="hero-title" className="hero-heading">
                Plan each inspection{' '}
                <span className="wine-text">with the map in view.</span>
              </h1>
              <p className="hero-copy">
                A shared view for county building teams.
              </p>
              <div className="hero-actions">
                <button
                  type="button"
                  className="mjengo-btn"
                  onClick={handleEnterWorkspace}
                >
                  Explore the map &nbsp;↗
                </button>
                <a
                  href="#workflow"
                  className="mjengo-btn alt"
                >
                  See the workflow &nbsp;↓
                </a>
              </div>
            </div>

            {/* Hero Map Visual with Sequenced Entrance and Replay Control */}
            <div
              id="map-preview"
              ref={heroVisualRef}
              key={motionKey}
              className="hero-visual"
              role="region"
              aria-label="Map illustration"
            >
              <div className="map-window">
                <svg
                  className="map-art"
                  viewBox="0 0 800 500"
                  preserveAspectRatio="xMidYMid slice"
                  aria-label="Illustrative map canvas, not live building data"
                  role="img"
                >
                  {/* Map Canvas Background */}
                  <rect width="800" height="500" fill="var(--brand-cream)" />

                  {/* Primary Road Arteries */}
                  <path
                    d="M-30 65 180 125 310 55 470 94 620 15 830 67M-20 195 165 150 265 215 450 175 600 223 820 165M-20 350 160 285 310 335 455 285 630 340 830 290M-20 470 180 420 320 455 500 410 670 460 830 414"
                    fill="none"
                    stroke="var(--brand-white)"
                    strokeWidth="24"
                  />
                  <path
                    d="M-30 65 180 125 310 55 470 94 620 15 830 67M-20 195 165 150 265 215 450 175 600 223 820 165M-20 350 160 285 310 335 455 285 630 340 830 290M-20 470 180 420 320 455 500 410 670 460 830 414"
                    fill="none"
                    stroke="var(--brand-border)"
                    strokeWidth="1.5"
                  />

                  {/* Secondary Streets */}
                  <path
                    d="M75 -20 130 520M280 -20 335 520M495 -20 450 520M690 -20 645 520M-20 260 820 250"
                    stroke="var(--brand-cream)"
                    strokeWidth="9"
                    fill="none"
                  />
                  <path
                    d="M75 -20 130 520M280 -20 335 520M495 -20 450 520M690 -20 645 520M-20 260 820 250"
                    stroke="var(--brand-white)"
                    strokeWidth="1"
                    fill="none"
                  />

                  {/* Restrained Building Footprints */}
                  <g fill="var(--brand-cream)" stroke="var(--brand-border)" strokeWidth="1">
                    <path d="m180 170 40-12 24 22-9 26-38 5-25-19z" />
                    <path d="M350 104l34-8 22 18-9 25-33 6-22-17z" />
                    <path d="M527 118l43-12 15 24-13 22-34 1-18-15z" />
                    <path d="M189 302l41-14 23 20-9 28-43 7-21-18z" />
                    <path d="M356 222l40-12 19 22-13 27-39 1-17-18z" />
                    <path d="M516 278l39-11 24 22-12 27-36 5-20-21z" />
                    <path d="M690 228l38-10 20 20-10 28-39 4-19-19z" />
                    <path d="M290 390l45-14 22 24-12 27-39 3-22-20z" />
                    <path d="M577 394l41-10 19 21-11 29-36 3-19-20z" />
                  </g>

                  {/* Supporting Blocks */}
                  <g fill="var(--brand-cream-light)">
                    <path d="m245 93 24-5 13 11-8 18-23 5-12-12z" />
                    <path d="M431 350l27-7 11 14-9 17-25 3-12-14z" />
                    <path d="M85 295l26-9 12 14-7 17-25 3-12-12z" />
                    <path d="M645 72l25-6 13 13-9 16-22 2-12-11z" />
                    <path d="M426 434l28-7 14 15-10 18-26 4-13-15z" />
                  </g>

                  {/* Sub-county Place Labels */}
                  <g
                    fontFamily="DM Sans, -apple-system, sans-serif"
                    fill="var(--brand-muted)"
                    fontSize="11"
                    letterSpacing="1.2"
                    fontWeight="600"
                  >
                    <text x="93" y="242">NAIROBI</text>
                    <text x="510" y="205">WESTLANDS</text>
                    <text x="318" y="373">KILIMANI</text>
                    <text x="663" y="321">PARKLANDS</text>
                  </g>

                  {/* Restrained Map Waypoints */}
                  <g fill="var(--brand-white)" stroke="var(--brand-maroon)" strokeWidth="2.2">
                    <circle cx="220" cy="222" r="6" />
                    <circle cx="390" cy="160" r="6" />
                    <circle cx="555" cy="322" r="6" />
                    <circle cx="635" cy="185" r="6" />
                  </g>

                  {/* Single Decorative Map Line Animation */}
                  <g
                    fill="none"
                    stroke="var(--brand-maroon)"
                    strokeWidth="2"
                    strokeDasharray="6 6"
                    opacity="0.65"
                  >
                    <path d="M220 222 Q300 160 390 160 T555 322" />
                  </g>
                </svg>

                {/* Map Tag */}
                <div className="map-tag">
                  <b>MAP PREVIEW</b>
                </div>

                {/* Map Toolbar with Replay Control */}
                <div className="map-toolbar" role="toolbar" aria-label="Map preview tools">
                  <button
                    type="button"
                    aria-label="Zoom in on preview"
                    onClick={() => {}}
                  >
                    ＋
                  </button>
                  <button
                    type="button"
                    aria-label="Zoom out on preview"
                    onClick={() => {}}
                  >
                    −
                  </button>
                  <button
                    type="button"
                    className="motion-replay-btn"
                    onClick={handleReplayMotion}
                    aria-label="Replay map entrance motion"
                  >
                    <span aria-hidden="true">↻</span> Replay
                  </button>
                </div>

                {/* Map Legend */}
                <div className="map-legend" role="complementary" aria-label="Map legend">
                  <strong>Map layers</strong>
                  <div className="legend-row">
                    <i className="swatch" aria-hidden="true" />
                    <span>Building footprint layer</span>
                  </div>
                  <div className="legend-row">
                    <i className="swatch line" aria-hidden="true" />
                    <span>Streets and place labels</span>
                  </div>
                </div>

                {/* Honest Map Disclaimer */}
                <div className="map-disclaimer">
                  Illustration. Replace with your MapTiler map and approved data.
                </div>
              </div>
            </div>
          </section>

          {/* Workflow Section Explained Once in Four Steps */}
          <section id="workflow" className="workflow-section" aria-labelledby="workflow-title">
            <div>
              <span className="hero-eyebrow">How teams work</span>
              <h2 id="workflow-title">From search to inspection report.</h2>
              <p style={{ color: 'var(--brand-maroon)', lineHeight: 1.7, fontSize: '13px', maxWidth: '390px' }}>
                Keep each step attached to the building record available to your team.
              </p>
            </div>

            <div className="workflow-steps">
              <div className="workflow-step">
                <span className="step-number">01</span>
                <div>
                  <p>
                    <b>Locate</b>
                    Search the MapTiler map or an approved building layer.
                  </p>
                </div>
              </div>

              <div className="workflow-step">
                <span className="step-number">02</span>
                <div>
                  <p>
                    <b>Review</b>
                    Open available property information and history.
                  </p>
                </div>
              </div>

              <div className="workflow-step">
                <span className="step-number">03</span>
                <div>
                  <p>
                    <b>Inspect</b>
                    Record observations against the selected building.
                  </p>
                </div>
              </div>

              <div className="workflow-step">
                <span className="step-number">04</span>
                <div>
                  <p>
                    <b>Report</b>
                    Review results from submitted inspection records.
                  </p>
                </div>
              </div>
            </div>
          </section>
        </main>

        {/* Clean Footer */}
        <footer className="landing-footer" role="contentinfo">
          <span>&copy; Mjengo Fiti</span>
          <span>Nairobi County Building Intelligence</span>
        </footer>
      </div>
    </div>
  );
};

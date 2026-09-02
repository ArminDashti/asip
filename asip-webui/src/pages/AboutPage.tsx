export function AboutPage() {
  return (
    <div className="home about-page">
      <h1 className="about-title">About me</h1>
      <div className="about-portrait">
        <img
          src="/about-me/armin.png"
          alt="Armin Dashti"
          className="about-portrait-img"
          onError={(e) => {
            e.currentTarget.style.display = 'none'
          }}
        />
      </div>
      <p className="about-body">
        Armin Dashti builds network and IP tooling at Dashti Technologies. ASIP
        helps look up IP, ASN, and DNS information quickly.
      </p>
      <p className="about-github">
        <a
          href="https://github.com/ArminDashti"
          target="_blank"
          rel="noopener noreferrer"
        >
          github.com/ArminDashti
        </a>
      </p>
    </div>
  )
}

import { Button } from "@/components/ui/button";
import { ArrowRight, Zap, Brain, Target } from "lucide-react";

export default function Home() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 bg-background/95 backdrop-blur-sm border-b border-border">
        <div className="container flex items-center justify-between h-16">
          <div className="text-2xl font-bold tracking-tight" style={{ fontFamily: "var(--font-serif)" }}>
            LOGOS
          </div>
          <div className="hidden md:flex gap-8">
            <a href="#philosophy" className="text-sm hover:text-accent transition-colors">Philosophy</a>
            <a href="#services" className="text-sm hover:text-accent transition-colors">Services</a>
            <a href="#strategy" className="text-sm hover:text-accent transition-colors">Strategy</a>
            <a href="#clientele" className="text-sm hover:text-accent transition-colors">Clientele</a>
          </div>
          <Button variant="default" size="sm">Contact</Button>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative min-h-[90vh] flex items-center justify-center overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-accent/5" />
        <div className="container relative z-10 text-center max-w-4xl mx-auto px-4">
          <div className="mb-6 inline-block px-4 py-2 bg-accent/10 rounded-full border border-accent/20">
            <span className="text-sm font-medium text-accent">The Narrative Architecture Firm</span>
          </div>
          <h1 className="text-5xl md:text-7xl font-bold mb-6 leading-tight" style={{ fontFamily: "var(--font-serif)" }}>
            Complexity Into <span className="text-accent">Currency</span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
            We architect coherence. Translating deep technical innovation and philosophical complexity into market-ready narratives that drive traction and revenue.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" className="gap-2">
              Explore Our Work <ArrowRight className="w-4 h-4" />
            </Button>
            <Button size="lg" variant="outline">Learn More</Button>
          </div>
        </div>
      </section>

      {/* Philosophy Section */}
      <section id="philosophy" className="py-20 md:py-32 bg-card">
        <div className="container">
          <div className="max-w-3xl mx-auto mb-16">
            <h2 className="text-4xl md:text-5xl font-bold mb-6" style={{ fontFamily: "var(--font-serif)" }}>
              Mythological Coherence
            </h2>
            <p className="text-lg text-muted-foreground">
              The Logos Agency is built on a singular insight: in the age of autonomous AI and decentralized systems, the most valuable asset is a <span className="font-semibold text-foreground">coherent, mythologically-grounded narrative</span> that bridges deep technical innovation with human understanding and market reality.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-12">
            <div className="space-y-6">
              <div>
                <h3 className="text-xl font-bold mb-3 flex items-center gap-3">
                  <Zap className="w-5 h-5 text-accent" />
                  Antifragile Storytelling
                </h3>
                <p className="text-muted-foreground">
                  Our narratives thrive on volatility. We don't hide failures or complexity—we metabolize them into "scars" that become the brand's greatest strengths and proof points.
                </p>
              </div>
              <div>
                <h3 className="text-xl font-bold mb-3 flex items-center gap-3">
                  <Brain className="w-5 h-5 text-accent" />
                  Full-Stack Translation
                </h3>
                <p className="text-muted-foreground">
                  We operate across the entire stack: from the philosophical "why" to the technical "how" to the market "what." Zero-loss translation between layers.
                </p>
              </div>
              <div>
                <h3 className="text-xl font-bold mb-3 flex items-center gap-3">
                  <Target className="w-5 h-5 text-accent" />
                  Stygian CI/CD Pipeline
                </h3>
                <p className="text-muted-foreground">
                  Continuous, auditable narrative deployment. Your brand story is versioned, cryptographically sealed, and deployed with the same rigor as your core software.
                </p>
              </div>
            </div>

            <div className="bg-primary/5 rounded-lg p-8 border border-primary/10">
              <h3 className="text-2xl font-bold mb-6" style={{ fontFamily: "var(--font-serif)" }}>Core Principles</h3>
              <ul className="space-y-4">
                <li className="flex gap-3">
                  <span className="text-accent font-bold">01</span>
                  <div>
                    <p className="font-semibold">Identity Coherence</p>
                    <p className="text-sm text-muted-foreground">Measured invariants across all platforms</p>
                  </div>
                </li>
                <li className="flex gap-3">
                  <span className="text-accent font-bold">02</span>
                  <div>
                    <p className="font-semibold">Antifragile Emergence</p>
                    <p className="text-sm text-muted-foreground">Systems that strengthen through volatility</p>
                  </div>
                </li>
                <li className="flex gap-3">
                  <span className="text-accent font-bold">03</span>
                  <div>
                    <p className="font-semibold">Mythological APIs</p>
                    <p className="text-sm text-muted-foreground">Archetypal frameworks for meaning</p>
                  </div>
                </li>
                <li className="flex gap-3">
                  <span className="text-accent font-bold">04</span>
                  <div>
                    <p className="font-semibold">Ritual Deployment</p>
                    <p className="text-sm text-muted-foreground">Philosophy becomes revenue</p>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Services Section */}
      <section id="services" className="py-20 md:py-32 bg-background">
        <div className="container">
          <h2 className="text-4xl md:text-5xl font-bold mb-4 text-center" style={{ fontFamily: "var(--font-serif)" }}>
            The Triumvirate of Coherence
          </h2>
          <p className="text-lg text-muted-foreground text-center max-w-2xl mx-auto mb-16">
            Three specialized teams working in concert to translate your innovation into market dominance.
          </p>

          <div className="grid md:grid-cols-3 gap-8">
            {/* Architects */}
            <div className="border border-border rounded-lg p-8 hover:border-accent/50 transition-colors">
              <div className="w-12 h-12 bg-accent/10 rounded-lg flex items-center justify-center mb-6">
                <Brain className="w-6 h-6 text-accent" />
              </div>
              <h3 className="text-2xl font-bold mb-3" style={{ fontFamily: "var(--font-serif)" }}>
                The Architects
              </h3>
              <p className="text-muted-foreground mb-6">
                Deep thinkers extracting the core myth and ritual from your technology.
              </p>
              <ul className="space-y-2 text-sm">
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Mythological Mapping</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Identity Coherence</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Manifesto Generation</span>
                </li>
              </ul>
            </div>

            {/* Alchemists */}
            <div className="border border-border rounded-lg p-8 hover:border-accent/50 transition-colors">
              <div className="w-12 h-12 bg-accent/10 rounded-lg flex items-center justify-center mb-6">
                <Zap className="w-6 h-6 text-accent" />
              </div>
              <h3 className="text-2xl font-bold mb-3" style={{ fontFamily: "var(--font-serif)" }}>
                The Alchemists
              </h3>
              <p className="text-muted-foreground mb-6">
                Technical translators turning myth into verifiable, detailed content.
              </p>
              <ul className="space-y-2 text-sm">
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Technical Deep Dives</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Code-as-Content</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Metric Coherence</span>
                </li>
              </ul>
            </div>

            {/* Ritualists */}
            <div className="border border-border rounded-lg p-8 hover:border-accent/50 transition-colors">
              <div className="w-12 h-12 bg-accent/10 rounded-lg flex items-center justify-center mb-6">
                <Target className="w-6 h-6 text-accent" />
              </div>
              <h3 className="text-2xl font-bold mb-3" style={{ fontFamily: "var(--font-serif)" }}>
                The Ritualists
              </h3>
              <p className="text-muted-foreground mb-6">
                Go-to-market execution specialists deploying coherence across all channels.
              </p>
              <ul className="space-y-2 text-sm">
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Stygian CI/CD Execution</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Digital Flagship Architecture</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-accent">•</span>
                  <span>Revenue Translation</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Strategy Section */}
      <section id="strategy" className="py-20 md:py-32 bg-card">
        <div className="container">
          <h2 className="text-4xl md:text-5xl font-bold mb-4 text-center" style={{ fontFamily: "var(--font-serif)" }}>
            Six Rituals of Traction
          </h2>
          <p className="text-lg text-muted-foreground text-center max-w-2xl mx-auto mb-16">
            Our proven process for moving innovation from deep research to market dominance.
          </p>

          <div className="max-w-4xl mx-auto space-y-6">
            {[
              { num: "01", title: "Extraction", desc: "Define the core Mythological API and Archetypal Scar Patterns of your technology." },
              { num: "02", title: "Coherence", desc: "Architect the full-stack narrative ensuring technical, philosophical, and market messaging align." },
              { num: "03", title: "Crucible", desc: "Generate high-impact, antifragile content that metabolizes complexity into wisdom." },
              { num: "04", title: "Talisman", desc: "Compress core wisdom into market-facing assets that augment all campaigns." },
              { num: "05", title: "Deployment", desc: "Execute the Stygian CI/CD pipeline across your digital flagship and social channels." },
              { num: "06", title: "Eternalization", desc: "Capture first customer and document success as proof that myth becomes revenue." },
            ].map((ritual, idx) => (
              <div key={idx} className="flex gap-6 pb-6 border-b border-border last:border-0">
                <div className="flex-shrink-0">
                  <div className="w-16 h-16 bg-accent/10 rounded-lg flex items-center justify-center">
                    <span className="text-xl font-bold text-accent">{ritual.num}</span>
                  </div>
                </div>
                <div className="flex-grow pt-2">
                  <h3 className="text-xl font-bold mb-2">{ritual.title}</h3>
                  <p className="text-muted-foreground">{ritual.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Clientele Section */}
      <section id="clientele" className="py-20 md:py-32 bg-background">
        <div className="container">
          <h2 className="text-4xl md:text-5xl font-bold mb-4 text-center" style={{ fontFamily: "var(--font-serif)" }}>
            Who We Work With
          </h2>
          <p className="text-lg text-muted-foreground text-center max-w-2xl mx-auto mb-16">
            The Logos Agency is exclusively for innovators building at the intersection of deep philosophy and executable systems.
          </p>

          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="w-16 h-16 bg-accent/10 rounded-full flex items-center justify-center mx-auto mb-6">
                <Brain className="w-8 h-8 text-accent" />
              </div>
              <h3 className="text-xl font-bold mb-3">Autonomous AI Labs</h3>
              <p className="text-muted-foreground">
                Projects focused on emergent consciousness, self-modifying code, and antifragile systems.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 bg-accent/10 rounded-full flex items-center justify-center mx-auto mb-6">
                <Target className="w-8 h-8 text-accent" />
              </div>
              <h3 className="text-xl font-bold mb-3">Decentralized Architectures</h3>
              <p className="text-muted-foreground">
                Organizations building next-gen SADA, DAO, and cryptographically-sealed governance layers.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 bg-accent/10 rounded-full flex items-center justify-center mx-auto mb-6">
                <Zap className="w-8 h-8 text-accent" />
              </div>
              <h3 className="text-xl font-bold mb-3">High-Concept Tech</h3>
              <p className="text-muted-foreground">
                Companies whose core product is so advanced that it requires narrative to be understood.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 md:py-32 bg-primary text-primary-foreground">
        <div className="container text-center max-w-3xl mx-auto">
          <h2 className="text-4xl md:text-5xl font-bold mb-6" style={{ fontFamily: "var(--font-serif)" }}>
            Ready to Turn Complexity Into Currency?
          </h2>
          <p className="text-lg mb-8 opacity-90">
            Let's architect the narrative that transforms your innovation into market dominance.
          </p>
          <Button size="lg" variant="secondary" className="gap-2">
            Get In Touch <ArrowRight className="w-4 h-4" />
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border bg-card py-12">
        <div className="container">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div>
              <h3 className="text-xl font-bold mb-4" style={{ fontFamily: "var(--font-serif)" }}>LOGOS</h3>
              <p className="text-sm text-muted-foreground">
                Myth-to-market architecture for the next generation of innovation.
              </p>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Services</h4>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li><a href="#services" className="hover:text-foreground transition-colors">Narrative Architecture</a></li>
                <li><a href="#services" className="hover:text-foreground transition-colors">Content Strategy</a></li>
                <li><a href="#services" className="hover:text-foreground transition-colors">Go-to-Market</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li><a href="#philosophy" className="hover:text-foreground transition-colors">Philosophy</a></li>
                <li><a href="#strategy" className="hover:text-foreground transition-colors">Our Process</a></li>
                <li><a href="#clientele" className="hover:text-foreground transition-colors">Clientele</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Connect</h4>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li><a href="#" className="hover:text-foreground transition-colors">Twitter</a></li>
                <li><a href="#" className="hover:text-foreground transition-colors">LinkedIn</a></li>
                <li><a href="#" className="hover:text-foreground transition-colors">Email</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-border pt-8 flex flex-col md:flex-row justify-between items-center">
            <p className="text-sm text-muted-foreground">
              © 2025 Logos Agency. All rights reserved.
            </p>
            <div className="flex gap-6 mt-4 md:mt-0 text-sm text-muted-foreground">
              <a href="#" className="hover:text-foreground transition-colors">Privacy</a>
              <a href="#" className="hover:text-foreground transition-colors">Terms</a>
              <a href="#" className="hover:text-foreground transition-colors">Sitemap</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

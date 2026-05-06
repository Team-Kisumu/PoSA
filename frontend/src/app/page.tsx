import Link from "next/link";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

// Pipeline step component.
function Step({ number, title, description }: { number: string; title: string; description: string }) {
  return (
    <div className="relative flex flex-col items-center text-center">
      <div className="w-12 h-12 rounded-full bg-black dark:bg-white flex items-center justify-center mb-4">
        <span className="text-white dark:text-black font-bold">{number}</span>
      </div>
      <h3 className="font-semibold text-black dark:text-white mb-2">{title}</h3>
      <p className="text-sm text-zinc-500 leading-relaxed">{description}</p>
    </div>
  );
}

// Feature card component.
function Feature({ title, description, tag }: { title: string; description: string; tag: string }) {
  return (
    <div className="p-6 rounded-xl border border-zinc-200 dark:border-zinc-800 hover:border-zinc-400 dark:hover:border-zinc-600 transition-colors">
      <span className="inline-block px-2.5 py-1 text-xs font-medium rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 mb-4">
        {tag}
      </span>
      <h3 className="font-semibold text-black dark:text-white mb-2">{title}</h3>
      <p className="text-sm text-zinc-500 leading-relaxed">{description}</p>
    </div>
  );
}

// Stat component.
function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="text-center">
      <p className="text-3xl font-bold text-black dark:text-white">{value}</p>
      <p className="text-sm text-zinc-500 mt-1">{label}</p>
    </div>
  );
}

export default function Home() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-black flex flex-col">
      <Navbar />

      <main className="flex-1">
        {/* Hero */}
        <section className="max-w-6xl mx-auto px-6 pt-20 pb-16 text-center">
          <div className="inline-block mb-8">
            <img src="/logo.png" alt="PoSA" className="h-16 mx-auto" />
          </div>
          <div className="inline-block px-4 py-1.5 rounded-full border border-zinc-200 dark:border-zinc-800 text-xs text-zinc-500 mb-8">
            Live on Flow Testnet
          </div>
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-black dark:text-white leading-tight mb-6">
            AI-Verified Work.
            <br />
            <span className="text-zinc-400">Blockchain-Trusted Proof.</span>
          </h1>
          <p className="text-lg text-zinc-600 dark:text-zinc-400 max-w-2xl mx-auto mb-10 leading-relaxed">
            PoSA uses artificial intelligence to evaluate your code and writing,
            then issues tamper-proof, verifiable credentials anchored on the blockchain.
            Prove your skills with cryptographic certainty.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/analyze"
              className="px-8 py-3.5 bg-black text-white rounded-lg font-medium hover:bg-zinc-800 transition-colors dark:bg-white dark:text-black dark:hover:bg-zinc-200"
            >
              Start Analyzing
            </Link>
            <Link
              href="/verify"
              className="px-8 py-3.5 border border-zinc-300 dark:border-zinc-700 rounded-lg font-medium text-zinc-700 dark:text-zinc-300 hover:border-zinc-500 dark:hover:border-zinc-500 transition-colors"
            >
              Verify a Credential
            </Link>
          </div>
        </section>

        {/* Stats */}
        <section className="border-y border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950">
          <div className="max-w-4xl mx-auto px-6 py-12 grid grid-cols-2 md:grid-cols-4 gap-8">
            <Stat value="25" label="Languages Supported" />
            <Stat value="176+" label="Analysis Rules" />
            <Stat value="4" label="Storage Backends" />
            <Stat value="0" label="Critical Vulnerabilities" />
          </div>
        </section>

        {/* How It Works */}
        <section className="max-w-6xl mx-auto px-6 py-20">
          <div className="text-center mb-14">
            <h2 className="text-3xl font-bold text-black dark:text-white mb-4">How It Works</h2>
            <p className="text-zinc-500 max-w-xl mx-auto">
              From submission to verifiable proof in four steps.
              Every credential is independently verifiable by anyone.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8">
            <Step
              number="1"
              title="Submit"
              description="Upload a file, paste code, or enter a GitHub repository URL for analysis."
            />
            <Step
              number="2"
              title="Analyze"
              description="AI evaluates your work for security, quality, logic, and style across 25 languages."
            />
            <Step
              number="3"
              title="Store"
              description="The evaluation report is stored on IPFS/Filecoin via Lighthouse for permanent access."
            />
            <Step
              number="4"
              title="Anchor"
              description="A cryptographic hash is recorded on the Flow blockchain and a verifiable credential is minted."
            />
          </div>
          {/* Connector line (desktop only) */}
          <div className="hidden lg:block relative -mt-[7.5rem] mb-16 mx-auto" style={{ width: "75%", height: "2px" }}>
            <div className="absolute inset-0 bg-gradient-to-r from-zinc-300 via-zinc-400 to-zinc-300 dark:from-zinc-700 dark:via-zinc-600 dark:to-zinc-700" />
          </div>
        </section>

        {/* Features */}
        <section className="bg-white dark:bg-zinc-950 border-y border-zinc-200 dark:border-zinc-800">
          <div className="max-w-6xl mx-auto px-6 py-20">
            <div className="text-center mb-14">
              <h2 className="text-3xl font-bold text-black dark:text-white mb-4">Built for Trust</h2>
              <p className="text-zinc-500 max-w-xl mx-auto">
                Every layer is designed for security, transparency, and verifiability.
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <Feature
                tag="AI"
                title="Multi-Language Analysis"
                description="Pattern-based static analysis for 25 programming languages with 176+ rules covering security, quality, logic, and style."
              />
              <Feature
                tag="AI"
                title="Impulse AI Integration"
                description="Supplementary AI-powered evaluation via Impulse Labs SSE streaming API for deeper code understanding."
              />
              <Feature
                tag="Storage"
                title="Decentralized Persistence"
                description="Reports stored on IPFS/Filecoin via Lighthouse with CID-based content addressing and LRU caching."
              />
              <Feature
                tag="Blockchain"
                title="On-Chain Credentials"
                description="Tamper-proof credentials minted on Flow with admin-restricted access, duplicate prevention, and immutable records."
              />
              <Feature
                tag="Security"
                title="Defense in Depth"
                description="6-step upload validation, fuzz-tested endpoints, adversarial AI testing, and formally verified smart contract."
              />
              <Feature
                tag="Verification"
                title="Independent Verifiability"
                description="Anyone can verify a credential by checking the CID against the on-chain record and retrieving the report from IPFS."
              />
            </div>
          </div>
        </section>

        {/* Tech Stack */}
        <section className="max-w-6xl mx-auto px-6 py-20">
          <div className="text-center mb-14">
            <h2 className="text-3xl font-bold text-black dark:text-white mb-4">Technology Stack</h2>
            <p className="text-zinc-500 max-w-xl mx-auto">
              Purpose-built with modern, auditable technologies across every layer.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              { name: "Go", role: "Backend API", detail: "REST API, validation, orchestration" },
              { name: "Python", role: "AI Engine", detail: "FastAPI, pattern analysis, Impulse AI" },
              { name: "Next.js", role: "Frontend", detail: "React, TypeScript, Tailwind CSS" },
              { name: "Cadence", role: "Smart Contract", detail: "Flow blockchain, credential minting" },
              { name: "Lighthouse", role: "Primary Storage", detail: "IPFS/Filecoin pinning service" },
              { name: "Beryx", role: "Chain Queries", detail: "Filecoin RPC via Zondax" },
              { name: "Flow", role: "Blockchain", detail: "Testnet at 0xf8a2...75a1" },
              { name: "Impulse AI", role: "AI Analysis", detail: "SSE streaming code evaluation" },
            ].map((tech) => (
              <div key={tech.name} className="p-5 rounded-xl border border-zinc-200 dark:border-zinc-800">
                <p className="font-semibold text-black dark:text-white">{tech.name}</p>
                <p className="text-xs text-zinc-400 mt-0.5">{tech.role}</p>
                <p className="text-sm text-zinc-500 mt-2">{tech.detail}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Use Cases */}
        <section className="bg-white dark:bg-zinc-950 border-y border-zinc-200 dark:border-zinc-800">
          <div className="max-w-6xl mx-auto px-6 py-20">
            <div className="text-center mb-14">
              <h2 className="text-3xl font-bold text-black dark:text-white mb-4">Use Cases</h2>
              <p className="text-zinc-500 max-w-xl mx-auto">
                Verifiable proof of skill for the digital economy.
              </p>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {[
                { title: "Freelance Verification", description: "Prove delivered work quality to clients with AI-scored, blockchain-anchored credentials." },
                { title: "Developer Portfolios", description: "Back project claims with verifiable AI evaluation scores that anyone can independently check." },
                { title: "Security Audit Certification", description: "Immutable proof of audit completion with detailed findings stored on decentralized storage." },
                { title: "Academic Credentialing", description: "Tamper-proof assignment and thesis evaluations anchored on-chain for permanent verification." },
                { title: "Hiring Pipelines", description: "Replace unverifiable claims with cryptographic proof of real skills evaluated by AI." },
                { title: "Open Source Contributions", description: "Verifiable quality scores for contributions to open source projects and repositories." },
              ].map((uc) => (
                <div key={uc.title} className="p-6 rounded-xl border border-zinc-200 dark:border-zinc-800">
                  <h3 className="font-semibold text-black dark:text-white mb-2">{uc.title}</h3>
                  <p className="text-sm text-zinc-500 leading-relaxed">{uc.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="max-w-6xl mx-auto px-6 py-20 text-center">
          <h2 className="text-3xl font-bold text-black dark:text-white mb-4">
            Ready to prove your skills?
          </h2>
          <p className="text-zinc-500 max-w-lg mx-auto mb-8">
            Submit your code or repository for AI evaluation and receive
            a verifiable credential in minutes.
          </p>
          <Link
            href="/analyze"
            className="inline-block px-8 py-3.5 bg-black text-white rounded-lg font-medium hover:bg-zinc-800 transition-colors dark:bg-white dark:text-black dark:hover:bg-zinc-200"
          >
            Get Started
          </Link>
        </section>
      </main>

      <Footer />
    </div>
  );
}

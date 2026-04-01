import Link from "next/link";

export default function Footer() {
  return (
    <footer className="border-t border-zinc-200 dark:border-zinc-800 mt-auto">
      <div className="max-w-6xl mx-auto px-6 py-8">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-8 mb-8">
          <div>
            <h4 className="font-semibold text-black dark:text-white mb-3">PoSA</h4>
            <p className="text-sm text-zinc-500 leading-relaxed">
              AI-Verified Work. Blockchain-Trusted Proof.
              Decentralized credential system for the digital economy.
            </p>
          </div>
          <div>
            <h4 className="font-semibold text-black dark:text-white mb-3">Product</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="/analyze" className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">Analyze Code</Link></li>
              <li><Link href="/verify" className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">Verify Credential</Link></li>
              <li>
                <a href="https://testnet.flowscan.io/account/0xf8a2fcf3389475a1" target="_blank" rel="noopener noreferrer"
                  className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">
                  Smart Contract
                </a>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="font-semibold text-black dark:text-white mb-3">Resources</h4>
            <ul className="space-y-2 text-sm">
              <li>
                <a href="https://github.com/Team-Kisumu/PoSA" target="_blank" rel="noopener noreferrer"
                  className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">
                  GitHub
                </a>
              </li>
              <li>
                <a href="https://github.com/Team-Kisumu/PoSA/blob/versions/SECURITY.md" target="_blank" rel="noopener noreferrer"
                  className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">
                  Security Policy
                </a>
              </li>
              <li>
                <a href="https://github.com/Team-Kisumu/PoSA/blob/versions/LICENSE" target="_blank" rel="noopener noreferrer"
                  className="text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-300 transition-colors">
                  MIT License
                </a>
              </li>
            </ul>
          </div>
        </div>
        <div className="border-t border-zinc-200 dark:border-zinc-800 pt-6 flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-xs text-zinc-400">PoSA {"\u00A9"} 2026 Joel Amos. All rights reserved.</p>
          <div className="flex items-center gap-4 text-xs text-zinc-400">
            <span>Impulse AI</span>
            <span className="text-zinc-300 dark:text-zinc-700">{"\u00B7"}</span>
            <span>Filecoin</span>
            <span className="text-zinc-300 dark:text-zinc-700">{"\u00B7"}</span>
            <span>Flow</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

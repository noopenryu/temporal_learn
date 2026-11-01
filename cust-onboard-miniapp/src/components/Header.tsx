import React, { useEffect, useState } from 'react';
import { Wallet, Copy, Check, ShieldCheck, RefreshCw } from 'lucide-react';
import { onboardingApi } from '../api/onboardingApi';
import { OnboardingApplication } from '../types/onboarding';

interface HeaderProps {
  application?: OnboardingApplication | null;
  onReset: () => void;
}

export const Header: React.FC<HeaderProps> = ({ application, onReset }) => {
  const [backendOnline, setBackendOnline] = useState<boolean | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let mounted = true;
    const check = async () => {
      const ok = await onboardingApi.checkHealth();
      if (mounted) setBackendOnline(ok);
    };
    check();
    const interval = setInterval(check, 10000);
    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  const copyAppId = () => {
    if (!application?.id) return;
    navigator.clipboard.writeText(application.id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <header className="border-b border-slate-200 bg-white/90 backdrop-blur sticky top-0 z-40 shadow-xs">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
        {/* Brand */}
        <button
          type="button"
          onClick={onReset}
          className="flex items-center space-x-3 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 rounded-xl p-1 -ml-1 transition-opacity hover:opacity-90"
        >
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center shadow-md shadow-blue-500/20 text-white">
            <Wallet className="w-5 h-5" aria-hidden="true" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-lg tracking-tight text-slate-900">PayFlow</span>
              <span className="text-[11px] font-semibold px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 border border-blue-200/60">
                E-Wallet
              </span>
            </div>
            <p className="text-[11px] text-slate-500 font-medium">Digital Account Opening Gateway</p>
          </div>
        </button>

        {/* Right side info */}
        <div className="flex items-center space-x-3 sm:space-x-4">
          {/* Security Guarantee Pill */}
          <div className="hidden md:flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-50 border border-slate-200 text-xs text-slate-600">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" aria-hidden="true" />
            <span className="font-medium text-[11px]">Bank-Grade TLS</span>
          </div>

          {/* Gateway Status */}
          <div className="flex items-center space-x-1.5 px-2.5 py-1 rounded-full bg-slate-50 border border-slate-200 text-xs">
            {backendOnline === true && (
              <>
                <span className="w-2 h-2 rounded-full bg-emerald-500" aria-hidden="true" />
                <span className="text-slate-600 text-[11px] font-medium hidden sm:inline">Gateway Online</span>
              </>
            )}
            {backendOnline === false && (
              <>
                <span className="w-2 h-2 rounded-full bg-rose-500" aria-hidden="true" />
                <span className="text-rose-600 text-[11px] font-medium">Gateway Offline</span>
              </>
            )}
            {backendOnline === null && (
              <>
                <span className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" aria-hidden="true" />
                <span className="text-slate-500 text-[11px]">Connecting...</span>
              </>
            )}
          </div>

          {/* Active Application Identifier Pill */}
          {application && (
            <div className="hidden sm:flex items-center space-x-2 px-2.5 py-1 rounded-lg bg-slate-100 border border-slate-200 text-xs text-slate-700 font-mono">
              <span className="text-slate-500 text-[11px]">App:</span>
              <span className="font-semibold">{application.id.slice(0, 8)}…</span>
              <button
                type="button"
                onClick={copyAppId}
                aria-label="Copy full application ID"
                title="Copy full UUID"
                className="hover:text-blue-600 transition-colors p-1 rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-600" aria-hidden="true" /> : <Copy className="w-3.5 h-3.5" aria-hidden="true" />}
              </button>
            </div>
          )}

          {/* New / Reset */}
          {application && (
            <button
              type="button"
              onClick={onReset}
              className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-slate-200 bg-white text-slate-700 text-xs font-semibold hover:bg-slate-50 hover:text-slate-900 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 shadow-2xs"
            >
              <RefreshCw className="w-3.5 h-3.5 text-slate-500" aria-hidden="true" />
              <span className="hidden sm:inline">Start Over</span>
            </button>
          )}
        </div>
      </div>
    </header>
  );
};

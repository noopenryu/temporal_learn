import React, { useState } from 'react';
import {
  CheckCircle2,
  Copy,
  Check,
  ArrowRight,
  QrCode,
  RotateCw,
  Wallet,
} from 'lucide-react';
import { OnboardingApplication } from '../../types/onboarding';

interface WalletSuccessCardProps {
  application: OnboardingApplication;
  onNewSession: () => void;
}

export const WalletSuccessCard: React.FC<WalletSuccessCardProps> = ({
  application,
  onNewSession,
}) => {
  const [copied, setCopied] = useState(false);
  const [isFlipped, setIsFlipped] = useState(false);
  const [showQrModal, setShowQrModal] = useState(false);

  const fullName = `${application.personal_info.first_name} ${application.personal_info.last_name}`;
  const accountId = application.wallet_account_id || `WAL-${application.id.slice(0, 8).toUpperCase()}`;
  const tier = application.approved_tier || application.wallet_tier_requested || 'TIER_1_BASIC';
  const isPremium = tier === 'TIER_2_PREMIUM';

  const copyAccount = () => {
    navigator.clipboard.writeText(accountId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="max-w-xl mx-auto space-y-6">
      {/* Celebration Header */}
      <div className="text-center space-y-2">
        <div className="inline-flex p-3 rounded-2xl bg-emerald-100 text-emerald-700 shadow-md shadow-emerald-500/10 mb-1">
          <CheckCircle2 className="w-8 h-8" />
        </div>
        <h2 className="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
          Welcome to PayFlow E-Wallet!
        </h2>
        <p className="text-xs sm:text-sm text-slate-500 max-w-md mx-auto leading-relaxed">
          Your customer identity is verified and your digital multi-currency wallet account has been provisioned.
        </p>
      </div>

      {/* Interactive 3D Virtual Wallet Card */}
      <div className="perspective-1000">
        <div
          onClick={() => setIsFlipped(!isFlipped)}
          className={`relative w-full aspect-1.58/1 rounded-3xl p-6 sm:p-7 shadow-2xl cursor-pointer transition-transform duration-700 transform-style-3d select-none ${
            isFlipped ? 'rotate-y-180' : ''
          } ${
            isPremium
              ? 'bg-gradient-to-tr from-slate-900 via-indigo-950 to-slate-950 text-white border border-indigo-500/30'
              : 'bg-gradient-to-tr from-blue-700 via-blue-800 to-indigo-900 text-white border border-blue-400/30'
          }`}
        >
          {/* Card Front */}
          <div className="absolute inset-0 p-6 sm:p-7 flex flex-col justify-between backface-hidden">
            {/* Top Row: Brand & Tier Badge */}
            <div className="flex justify-between items-start">
              <div className="flex items-center space-x-2">
                <div className="w-8 h-8 rounded-lg bg-white/10 backdrop-blur border border-white/20 flex items-center justify-center">
                  <Wallet className="w-4 h-4 text-white" />
                </div>
                <div>
                  <span className="font-extrabold tracking-tight text-sm">PayFlow</span>
                  <span className="block text-[9px] uppercase tracking-widest text-slate-300">Digital Wallet</span>
                </div>
              </div>

              <span className={`text-[10px] font-bold px-2.5 py-0.5 rounded-full border ${
                isPremium
                  ? 'bg-amber-400/20 text-amber-300 border-amber-400/30'
                  : 'bg-white/20 text-white border-white/30'
              }`}>
                {isPremium ? 'TIER 2 • PREMIUM' : 'TIER 1 • BASIC'}
              </span>
            </div>

            {/* Middle Row: Chip & NFC */}
            <div className="flex items-center space-x-3 my-2">
              <div className="w-10 h-7 rounded-md bg-gradient-to-r from-amber-200 to-amber-400 border border-amber-500/60 shadow-xs flex items-center justify-around px-1">
                <div className="w-2 h-4 border-r border-amber-600/40" />
                <div className="w-2 h-4 border-l border-amber-600/40" />
              </div>
              <span className="text-[10px] text-white/70 font-mono tracking-widest">NFC ENABLED</span>
            </div>

            {/* Bottom Section: Account Number, Name & Flip hint */}
            <div className="space-y-3">
              <div>
                <span className="text-[9px] uppercase tracking-widest text-slate-300 block">Wallet Account Identifier</span>
                <span className="text-lg sm:text-xl font-mono font-bold tracking-widest text-white">
                  {accountId}
                </span>
              </div>

              <div className="flex justify-between items-end border-t border-white/15 pt-2 text-xs">
                <div>
                  <p className="text-[9px] text-slate-300 uppercase tracking-wider">Account Holder</p>
                  <p className="font-semibold text-white tracking-wide">{fullName}</p>
                </div>
                <div className="flex items-center space-x-1 text-[10px] text-slate-300">
                  <RotateCw className="w-3 h-3" />
                  <span>Click to flip</span>
                </div>
              </div>
            </div>
          </div>

          {/* Card Back */}
          <div className="absolute inset-0 p-6 sm:p-7 flex flex-col justify-between backface-hidden rotate-y-180 bg-slate-900 border border-slate-700/60 rounded-3xl">
            <div className="w-full h-10 bg-slate-950 -mx-6 sm:-mx-7 mt-2" />
            <div className="flex items-center justify-between px-2">
              <div className="text-[10px] text-slate-400 font-mono">
                CUSTOMER SERVICE: +1 800 PAYFLOW
              </div>
              <div className="bg-white text-slate-900 font-mono text-xs font-bold px-3 py-1 rounded">
                CVV 882
              </div>
            </div>
            <p className="text-[9px] text-slate-400 leading-tight">
              This card is a digital e-wallet instrument issued by PayFlow. Subject to the Terms of Service. Click again to flip front.
            </p>
          </div>
        </div>
      </div>

      {/* Account Details & Quick Actions */}
      <div className="bg-white border border-slate-200/90 rounded-2xl p-5 shadow-xs space-y-4">
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-xs">
          <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
            <span className="text-slate-500 block text-[10px] uppercase font-semibold">Account Status</span>
            <span className="font-bold text-emerald-700 mt-0.5 flex items-center">
              <CheckCircle2 className="w-3.5 h-3.5 mr-1 text-emerald-600" /> Active & Verified
            </span>
          </div>

          <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
            <span className="text-slate-500 block text-[10px] uppercase font-semibold">Daily Spending Cap</span>
            <span className="font-bold font-mono text-slate-900 mt-0.5 block">
              {isPremium ? '$10,000 USD' : '$1,000 USD'}
            </span>
          </div>

          <div className="p-3 rounded-xl bg-slate-50 border border-slate-100 col-span-2 sm:col-span-1">
            <span className="text-slate-500 block text-[10px] uppercase font-semibold">Storage Balance</span>
            <span className="font-bold font-mono text-slate-900 mt-0.5 block">
              ${isPremium ? '50,000' : '5,000'} USD Max
            </span>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-wrap items-center gap-2 pt-1">
          <button
            type="button"
            onClick={copyAccount}
            className="flex-1 py-2.5 px-4 rounded-xl border border-slate-200 bg-white hover:bg-slate-50 text-slate-800 text-xs font-semibold flex items-center justify-center space-x-1.5 transition-colors shadow-2xs"
          >
            {copied ? <Check className="w-4 h-4 text-emerald-600" /> : <Copy className="w-4 h-4 text-slate-600" />}
            <span>{copied ? 'Copied to Clipboard!' : 'Copy Account Number'}</span>
          </button>

          <button
            type="button"
            onClick={() => setShowQrModal(!showQrModal)}
            className="flex-1 py-2.5 px-4 rounded-xl border border-slate-200 bg-white hover:bg-slate-50 text-slate-800 text-xs font-semibold flex items-center justify-center space-x-1.5 transition-colors shadow-2xs"
          >
            <QrCode className="w-4 h-4 text-slate-600" />
            <span>Show QR Code</span>
          </button>

          <button
            type="button"
            onClick={onNewSession}
            className="w-full sm:w-auto py-2.5 px-5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold flex items-center justify-center space-x-1.5 transition-colors shadow-xs"
          >
            <span>Start Another Application</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* QR Code expansion */}
        {showQrModal && (
          <div className="p-4 rounded-xl bg-slate-50 border border-slate-200 text-center space-y-2">
            <div className="w-32 h-32 mx-auto bg-white p-2 rounded-xl border border-slate-200 flex items-center justify-center shadow-2xs">
              <div className="w-full h-full bg-slate-900 rounded flex items-center justify-center text-white font-mono text-[10px] p-2 text-center">
                QR://{accountId}
              </div>
            </div>
            <p className="text-xs font-semibold text-slate-700">Scan to Link Mobile App</p>
            <p className="text-[11px] text-slate-500">Scan with PayFlow App on iOS / Android to bind this wallet</p>
          </div>
        )}
      </div>
    </div>
  );
};

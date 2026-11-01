import React, { useState } from 'react';
import {
  UserPlus,
  LogIn,
  AlertCircle,
  Phone,
  Mail,
  Lock,
  Eye,
  EyeOff,
  Sparkles,
  ShieldCheck,
  Zap,
  Globe2,
  CheckCircle2,
} from 'lucide-react';
import { ResumeSessionRequest, StartSessionRequest } from '../../types/onboarding';

interface StartSessionStepProps {
  onStart: (req: StartSessionRequest) => Promise<void>;
  onResume: (req: ResumeSessionRequest) => Promise<void>;
  isLoading: boolean;
  error?: string | null;
  onClearError?: () => void;
}

export const StartSessionStep: React.FC<StartSessionStepProps> = ({
  onStart,
  onResume,
  isLoading,
  error,
  onClearError,
}) => {
  const [activeTab, setActiveTab] = useState<'start' | 'resume'>('start');

  // Start Form State
  const [email, setEmail] = useState('alice.smith@example.com');
  const [phoneNumber, setPhoneNumber] = useState('+15551234567');
  const [password, setPassword] = useState('Secret123!');
  const [showPassword, setShowPassword] = useState(false);

  // Resume Form State
  const [resumePhone, setResumePhone] = useState('+15551234567');
  const [resumePassword, setResumePassword] = useState('');
  const [showResumePassword, setShowResumePassword] = useState(false);

  // Field validation errors
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const validateStartForm = (): boolean => {
    const errors: Record<string, string> = {};
    const trimmedPhone = phoneNumber.trim();
    const trimmedEmail = email.trim();

    if (!trimmedPhone) {
      errors.phone = 'Mobile phone number is required';
    } else if (!/^\+?[1-9]\d{7,14}$/.test(trimmedPhone.replace(/[\s-]/g, ''))) {
      errors.phone = 'Enter a valid international phone number (e.g. +1 555 123 4567)';
    }

    if (!trimmedEmail) {
      errors.email = 'Email address is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      errors.email = 'Enter a valid email address';
    }

    if (!password) {
      errors.password = 'Application password is required';
    } else if (password.length < 6) {
      errors.password = 'Password must be at least 6 characters long';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const validateResumeForm = (): boolean => {
    const errors: Record<string, string> = {};
    const trimmedPhone = resumePhone.trim();

    if (!trimmedPhone) {
      errors.resumePhone = 'Mobile phone number is required';
    } else if (!/^\+?[1-9]\d{7,14}$/.test(trimmedPhone.replace(/[\s-]/g, ''))) {
      errors.resumePhone = 'Enter a valid phone number';
    }

    if (!resumePassword) {
      errors.resumePassword = 'Password is required to resume application';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleStartSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (onClearError) onClearError();
    if (!validateStartForm()) return;

    onStart({
      email: email.trim(),
      phone_number: phoneNumber.trim(),
      password,
    });
  };

  const handleResumeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (onClearError) onClearError();
    if (!validateResumeForm()) return;

    onResume({
      phone_number: resumePhone.trim(),
      password: resumePassword,
    });
  };

  // Quick fill personas for development & testing
  const applyPersona = (personaEmail: string, personaPhone: string, personaPass: string) => {
    setEmail(personaEmail);
    setPhoneNumber(personaPhone);
    setPassword(personaPass);
    setResumePhone(personaPhone);
    setResumePassword(personaPass);
    setFormErrors({});
    if (onClearError) onClearError();
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Hero Welcome Card */}
      <div className="bg-gradient-to-br from-slate-900 via-indigo-950 to-slate-900 rounded-2xl p-6 sm:p-8 text-white shadow-xl relative overflow-hidden">
        {/* Subtle background glow */}
        <div className="absolute top-0 right-0 -mt-8 -mr-8 w-48 h-48 bg-blue-500/20 rounded-full blur-3xl pointer-events-none" />

        <div className="relative z-10 space-y-3">
          <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-blue-500/20 border border-blue-400/30 text-blue-300 text-xs font-semibold">
            <Sparkles className="w-3.5 h-3.5 text-blue-400" />
            <span>Next-Gen Internet Wallet</span>
          </div>

          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
            Open Your PayFlow E-Wallet in 2 Minutes
          </h1>
          <p className="text-slate-300 text-sm max-w-lg leading-relaxed">
            Multi-currency balances, instant QR code checkout, virtual card provisioning, and zero cross-border markup.
          </p>

          <div className="pt-2 grid grid-cols-1 sm:grid-cols-3 gap-2.5 text-xs text-slate-300 font-medium">
            <div className="flex items-center space-x-1.5">
              <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
              <span>Instant Verification</span>
            </div>
            <div className="flex items-center space-x-1.5">
              <Globe2 className="w-4 h-4 text-blue-400 shrink-0" />
              <span>Multi-Currency Ready</span>
            </div>
            <div className="flex items-center space-x-1.5">
              <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
              <span>Licensed & Encrypted</span>
            </div>
          </div>
        </div>
      </div>

      {/* Main Authentication / Session Card */}
      <div className="bg-white border border-slate-200/90 rounded-2xl shadow-sm p-6 sm:p-8">
        {/* Segmented Tab Switcher */}
        <div className="flex p-1 bg-slate-100 rounded-xl mb-6">
          <button
            type="button"
            onClick={() => {
              setActiveTab('start');
              if (onClearError) onClearError();
            }}
            className={`flex-1 flex items-center justify-center space-x-2 py-2.5 px-4 rounded-lg font-semibold text-xs sm:text-sm transition-all ${
              activeTab === 'start'
                ? 'bg-white text-blue-900 shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <UserPlus className="w-4 h-4" />
            <span>New Application</span>
          </button>
          <button
            type="button"
            onClick={() => {
              setActiveTab('resume');
              if (onClearError) onClearError();
            }}
            className={`flex-1 flex items-center justify-center space-x-2 py-2.5 px-4 rounded-lg font-semibold text-xs sm:text-sm transition-all ${
              activeTab === 'resume'
                ? 'bg-white text-blue-900 shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <LogIn className="w-4 h-4" />
            <span>Resume Application</span>
          </button>
        </div>

        {/* Global Error Banner */}
        {error && (
          <div className="mb-5 p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-rose-800 text-xs flex items-start space-x-2.5">
            <AlertCircle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="font-semibold">Unable to proceed</p>
              <p className="mt-0.5">{error}</p>
            </div>
          </div>
        )}

        {/* Start New Application Form */}
        {activeTab === 'start' && (
          <form onSubmit={handleStartSubmit} className="space-y-4">
            <div>
              <label
                htmlFor="phone-input"
                className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5"
              >
                Mobile Phone Number
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
                  <Phone className="w-4 h-4" />
                </div>
                <input
                  id="phone-input"
                  type="tel"
                  value={phoneNumber}
                  onChange={(e) => setPhoneNumber(e.target.value)}
                  placeholder="+1 555 123 4567"
                  className={`w-full pl-10 pr-4 py-2.5 bg-slate-50 border rounded-xl text-sm text-slate-900 placeholder:text-slate-400 transition-colors focus:bg-white ${
                    formErrors.phone ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
                  }`}
                />
              </div>
              {formErrors.phone && (
                <p className="text-xs text-rose-600 mt-1 font-medium">{formErrors.phone}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="email-input"
                className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5"
              >
                Email Address
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
                  <Mail className="w-4 h-4" />
                </div>
                <input
                  id="email-input"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="alice.smith@example.com"
                  className={`w-full pl-10 pr-4 py-2.5 bg-slate-50 border rounded-xl text-sm text-slate-900 placeholder:text-slate-400 transition-colors focus:bg-white ${
                    formErrors.email ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
                  }`}
                />
              </div>
              {formErrors.email && (
                <p className="text-xs text-rose-600 mt-1 font-medium">{formErrors.email}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="password-input"
                className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5"
              >
                Create Application Password
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
                  <Lock className="w-4 h-4" />
                </div>
                <input
                  id="password-input"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Minimum 6 characters"
                  className={`w-full pl-10 pr-10 py-2.5 bg-slate-50 border rounded-xl text-sm text-slate-900 placeholder:text-slate-400 transition-colors focus:bg-white ${
                    formErrors.password ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-600"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              {formErrors.password && (
                <p className="text-xs text-rose-600 mt-1 font-medium">{formErrors.password}</p>
              )}
              <p className="text-[11px] text-slate-500 mt-1">
                Used to safely save and resume your draft if you disconnect.
              </p>
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full mt-2 py-3 px-4 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm transition-all shadow-md shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center space-x-2"
            >
              {isLoading ? (
                <>
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>Creating Application...</span>
                </>
              ) : (
                <>
                  <Zap className="w-4 h-4" />
                  <span>Start Onboarding</span>
                </>
              )}
            </button>
          </form>
        )}

        {/* Resume In-Progress Application Form */}
        {activeTab === 'resume' && (
          <form onSubmit={handleResumeSubmit} className="space-y-4">
            <div>
              <label
                htmlFor="resume-phone-input"
                className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5"
              >
                Registered Phone Number
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
                  <Phone className="w-4 h-4" />
                </div>
                <input
                  id="resume-phone-input"
                  type="tel"
                  value={resumePhone}
                  onChange={(e) => setResumePhone(e.target.value)}
                  placeholder="+1 555 123 4567"
                  className={`w-full pl-10 pr-4 py-2.5 bg-slate-50 border rounded-xl text-sm text-slate-900 placeholder:text-slate-400 transition-colors focus:bg-white ${
                    formErrors.resumePhone ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
                  }`}
                />
              </div>
              {formErrors.resumePhone && (
                <p className="text-xs text-rose-600 mt-1 font-medium">{formErrors.resumePhone}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="resume-password-input"
                className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5"
              >
                Application Password
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
                  <Lock className="w-4 h-4" />
                </div>
                <input
                  id="resume-password-input"
                  type={showResumePassword ? 'text' : 'password'}
                  value={resumePassword}
                  onChange={(e) => setResumePassword(e.target.value)}
                  placeholder="Enter your application password"
                  className={`w-full pl-10 pr-10 py-2.5 bg-slate-50 border rounded-xl text-sm text-slate-900 placeholder:text-slate-400 transition-colors focus:bg-white ${
                    formErrors.resumePassword ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowResumePassword(!showResumePassword)}
                  aria-label={showResumePassword ? 'Hide password' : 'Show password'}
                  className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-600"
                >
                  {showResumePassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              {formErrors.resumePassword && (
                <p className="text-xs text-rose-600 mt-1 font-medium">{formErrors.resumePassword}</p>
              )}
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full mt-2 py-3 px-4 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm transition-all shadow-md shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center space-x-2"
            >
              {isLoading ? (
                <>
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>Resuming Application...</span>
                </>
              ) : (
                <>
                  <LogIn className="w-4 h-4" />
                  <span>Resume My Application</span>
                </>
              )}
            </button>
          </form>
        )}

        {/* Quick Test Persona Pre-fills for easy development and reviewer test */}
        <div className="mt-8 pt-5 border-t border-slate-100">
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2.5">
            Quick Test Personas
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
            <button
              type="button"
              onClick={() => applyPersona('linh.tran@techwallet.io', '+84988776611', 'Secret123!')}
              className="p-2.5 text-left border border-slate-200 rounded-xl hover:border-blue-400 hover:bg-blue-50/40 transition-colors"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Linh Tran</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800">Tier 2</span>
              </div>
              <p className="text-[10px] text-slate-500 mt-0.5">Instant Clear Pass</p>
            </button>

            <button
              type="button"
              onClick={() => applyPersona('minh.officer.review@test.internal', '+84912345699', 'Secret123!')}
              className="p-2.5 text-left border border-slate-200 rounded-xl hover:border-amber-400 hover:bg-amber-50/40 transition-colors"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Minh Le</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-amber-100 text-amber-800">Review</span>
              </div>
              <p className="text-[10px] text-slate-500 mt-0.5">Name Flagged Case</p>
            </button>

            <button
              type="button"
              onClick={() => applyPersona('sarah.jenkins@expatglobal.co', '+15559876543', 'Secret123!')}
              className="p-2.5 text-left border border-slate-200 rounded-xl hover:border-blue-400 hover:bg-blue-50/40 transition-colors"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Sarah Jenkins</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-purple-100 text-purple-800">Expat</span>
              </div>
              <p className="text-[10px] text-slate-500 mt-0.5">International Passport</p>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

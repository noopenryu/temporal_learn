import React, { useId, useState } from 'react';
import { User, ArrowRight, ArrowLeft, CheckCircle2, AlertTriangle } from 'lucide-react';
import { PersonalInfo } from '../../types/onboarding';

interface PersonalInfoStepProps {
  initialData?: PersonalInfo;
  onSave: (data: PersonalInfo) => Promise<void>;
  onBack: () => void;
  isLoading: boolean;
}

const COUNTRIES = [
  { code: 'VN', name: 'Vietnam', flag: '🇻🇳' },
  { code: 'SG', name: 'Singapore', flag: '🇸🇬' },
  { code: 'US', name: 'United States', flag: '🇺🇸' },
  { code: 'GB', name: 'United Kingdom', flag: '🇬🇧' },
  { code: 'ID', name: 'Indonesia', flag: '🇮🇩' },
  { code: 'MY', name: 'Malaysia', flag: '🇲🇾' },
  { code: 'TH', name: 'Thailand', flag: '🇹🇭' },
  { code: 'JP', name: 'Japan', flag: '🇯🇵' },
  { code: 'DE', name: 'Germany', flag: '🇩🇪' },
  { code: 'FR', name: 'France', flag: '🇫🇷' },
  { code: 'AU', name: 'Australia', flag: '🇦🇺' },
];

export const PersonalInfoStep: React.FC<PersonalInfoStepProps> = ({
  initialData,
  onSave,
  onBack,
  isLoading,
}) => {
  const firstNameId = useId();
  const lastNameId = useId();
  const dobId = useId();
  const genderId = useId();
  const nationalityId = useId();
  const phoneId = useId();
  const emailId = useId();

  // Compute today's ISO date string (YYYY-MM-DD)
  const today = new Date();
  const todayString = today.toISOString().split('T')[0];

  // Default adult DOB (18 years ago today)
  const adultDate = new Date(today.getFullYear() - 18, today.getMonth(), today.getDate());
  const defaultAdultString = adultDate.toISOString().split('T')[0];

  // Sanitize initial DOB: eliminate year 0001 (Go zero-value) or dates before 1900
  const getCleanDob = (raw?: string): string => {
    if (!raw) return defaultAdultString;
    const datePart = raw.split('T')[0];
    if (datePart.startsWith('0001') || datePart.startsWith('0000') || datePart.startsWith('1900-01-01')) {
      return defaultAdultString;
    }
    return datePart;
  };

  const [firstName, setFirstName] = useState(initialData?.first_name || 'Alice');
  const [lastName, setLastName] = useState(initialData?.last_name || 'Smith');
  const [dob, setDob] = useState(() => getCleanDob(initialData?.date_of_birth));
  const [gender, setGender] = useState(initialData?.gender || 'FEMALE');
  const [nationality, setNationality] = useState(initialData?.nationality || 'VN');
  const [phoneNumber, setPhoneNumber] = useState(initialData?.phone_number || '+15551234567');
  const [email, setEmail] = useState(initialData?.email || 'alice.smith@example.com');

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Age calculation
  const calculateAge = (birthDateString: string): { age: number; isAdult: boolean; isFuture: boolean } => {
    if (!birthDateString) return { age: 0, isAdult: false, isFuture: false };
    const bDate = new Date(birthDateString);
    if (isNaN(bDate.getTime())) return { age: 0, isAdult: false, isFuture: false };

    if (bDate > today) {
      return { age: 0, isAdult: false, isFuture: true };
    }

    let age = today.getFullYear() - bDate.getFullYear();
    const m = today.getMonth() - bDate.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < bDate.getDate())) {
      age--;
    }
    return { age, isAdult: age >= 18, isFuture: false };
  };

  const { age, isAdult, isFuture } = calculateAge(dob);

  const validate = (): boolean => {
    const errs: Record<string, string> = {};

    if (!firstName.trim() || firstName.trim().length < 2) {
      errs.firstName = 'First name must be at least 2 characters';
    }
    if (!lastName.trim() || lastName.trim().length < 2) {
      errs.lastName = 'Last name must be at least 2 characters';
    }

    if (!dob) {
      errs.dob = 'Date of birth is required';
    } else if (isFuture) {
      errs.dob = 'Date of birth cannot be in the future';
    } else if (!isAdult) {
      errs.dob = 'Applicant must be at least 18 years old to open an account';
    }

    const trimmedPhone = phoneNumber.trim().replace(/[\s-]/g, '');
    if (!trimmedPhone) {
      errs.phoneNumber = 'Phone number is required';
    } else if (!/^\+?[1-9]\d{7,14}$/.test(trimmedPhone)) {
      errs.phoneNumber = 'Please enter a valid international phone number';
    }

    const trimmedEmail = email.trim();
    if (!trimmedEmail) {
      errs.email = 'Email address is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      errs.email = 'Please enter a valid email address';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    onSave({
      first_name: firstName.trim(),
      last_name: lastName.trim(),
      date_of_birth: dob,
      gender,
      nationality,
      phone_number: phoneNumber.trim(),
      email: email.trim(),
    });
  };

  return (
    <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center space-x-3 pb-6 border-b border-slate-100">
        <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
          <User className="w-5 h-5" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight">Personal Profile</h2>
          <p className="text-xs text-slate-500">Provide legal identification details as shown on your government ID.</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="pt-6 space-y-5">
        {/* Name Fields */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={firstNameId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              First Name <span className="text-rose-500">*</span>
            </label>
            <input
              id={firstNameId}
              name="given-name"
              type="text"
              autoComplete="given-name"
              required
              value={firstName}
              onChange={(e) => {
                setFirstName(e.target.value);
                if (errors.firstName) setErrors((prev) => ({ ...prev, firstName: '' }));
              }}
              placeholder="e.g. Alice"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.firstName ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.firstName && (
              <p className="text-xs text-rose-600 mt-1 font-medium">{errors.firstName}</p>
            )}
          </div>

          <div>
            <label htmlFor={lastNameId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Last Name <span className="text-rose-500">*</span>
            </label>
            <input
              id={lastNameId}
              name="family-name"
              type="text"
              autoComplete="family-name"
              required
              value={lastName}
              onChange={(e) => {
                setLastName(e.target.value);
                if (errors.lastName) setErrors((prev) => ({ ...prev, lastName: '' }));
              }}
              placeholder="e.g. Smith"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.lastName ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.lastName && (
              <p className="text-xs text-rose-600 mt-1 font-medium">{errors.lastName}</p>
            )}
          </div>
        </div>

        {/* Date of Birth */}
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label htmlFor={dobId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider">
              Date of Birth <span className="text-rose-500">*</span>
            </label>
            {dob && !isFuture && (
              <span
                className={`text-[11px] font-semibold px-2.5 py-0.5 rounded-full inline-flex items-center space-x-1 ${
                  isAdult
                    ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                    : 'bg-rose-50 text-rose-700 border border-rose-200 font-bold'
                }`}
              >
                {isAdult ? (
                  <>
                    <CheckCircle2 className="w-3 h-3 text-emerald-600 mr-1" />
                    <span>Age {age} • Eligible</span>
                  </>
                ) : (
                  <>
                    <AlertTriangle className="w-3 h-3 text-rose-600 mr-1" />
                    <span>Underage &lt;18 ({age} years old)</span>
                  </>
                )}
              </span>
            )}
          </div>

          <div className="relative">
            <input
              id={dobId}
              name="bday"
              type="date"
              autoComplete="bday"
              required
              max={todayString}
              min="1920-01-01"
              value={dob}
              onChange={(e) => {
                setDob(e.target.value);
                if (errors.dob) setErrors((prev) => ({ ...prev, dob: '' }));
              }}
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.dob || (!isAdult && dob)
                  ? 'border-rose-400 focus:border-rose-500'
                  : 'border-slate-300 focus:border-blue-600'
              }`}
            />
          </div>

          {errors.dob && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.dob}</p>}
          {!isAdult && dob && !isFuture && (
            <p className="text-xs text-rose-600 mt-1 font-medium">
              Legal requirements mandate customers must be at least 18 years old to open an internet wallet account.
            </p>
          )}
        </div>

        {/* Gender & Nationality */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={genderId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Gender
            </label>
            <select
              id={genderId}
              name="sex"
              value={gender}
              onChange={(e) => setGender(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            >
              <option value="FEMALE">Female</option>
              <option value="MALE">Male</option>
              <option value="NON_BINARY">Non-Binary</option>
              <option value="PREFER_NOT_TO_SAY">Prefer not to say</option>
            </select>
          </div>

          <div>
            <label htmlFor={nationalityId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Nationality <span className="text-rose-500">*</span>
            </label>
            <select
              id={nationalityId}
              name="country"
              value={nationality}
              onChange={(e) => setNationality(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white focus:border-blue-600"
            >
              {COUNTRIES.map((c) => (
                <option key={c.code} value={c.code}>
                  {c.flag} {c.name} ({c.code})
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Phone & Email */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={phoneId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Phone Number <span className="text-rose-500">*</span>
            </label>
            <input
              id={phoneId}
              name="tel"
              type="tel"
              autoComplete="tel"
              required
              value={phoneNumber}
              onChange={(e) => {
                setPhoneNumber(e.target.value);
                if (errors.phoneNumber) setErrors((prev) => ({ ...prev, phoneNumber: '' }));
              }}
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 font-mono transition-colors focus:bg-white ${
                errors.phoneNumber ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.phoneNumber && (
              <p className="text-xs text-rose-600 mt-1 font-medium">{errors.phoneNumber}</p>
            )}
          </div>

          <div>
            <label htmlFor={emailId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Email Address <span className="text-rose-500">*</span>
            </label>
            <input
              id={emailId}
              name="email"
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                if (errors.email) setErrors((prev) => ({ ...prev, email: '' }));
              }}
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm text-slate-900 transition-colors focus:bg-white ${
                errors.email ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.email && (
              <p className="text-xs text-rose-600 mt-1 font-medium">{errors.email}</p>
            )}
          </div>
        </div>

        {/* Controls */}
        <div className="pt-6 flex items-center justify-between border-t border-slate-100">
          <button
            type="button"
            onClick={onBack}
            className="flex items-center space-x-2 px-4 py-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 font-semibold text-xs transition-colors"
          >
            <ArrowLeft className="w-4 h-4" aria-hidden="true" />
            <span>Back</span>
          </button>

          <button
            type="submit"
            disabled={isLoading || !isAdult}
            className="flex items-center space-x-2 px-6 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs shadow-md shadow-blue-500/20 transition-all disabled:opacity-50"
          >
            <span>{isLoading ? 'Saving...' : 'Continue to Address'}</span>
            <ArrowRight className="w-4 h-4" aria-hidden="true" />
          </button>
        </div>
      </form>
    </div>
  );
};

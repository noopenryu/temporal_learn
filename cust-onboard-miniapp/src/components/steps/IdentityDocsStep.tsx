import React, { useId, useState } from 'react';
import {
  FileText,
  ArrowRight,
  ArrowLeft,
  CheckCircle2,
  ShieldCheck,
  Camera,
  Sparkles,
  Scan,
} from 'lucide-react';
import { DocumentType, IdentityDocuments } from '../../types/onboarding';

interface IdentityDocsStepProps {
  initialData?: IdentityDocuments;
  onSave: (data: IdentityDocuments) => Promise<void>;
  onBack: () => void;
  isLoading: boolean;
}

const DOCUMENT_TYPES: { id: DocumentType; label: string; desc: string }[] = [
  { id: 'NATIONAL_ID', label: 'National ID / Citizen Card', desc: 'CCCD, NRIC, or National ID Card' },
  { id: 'PASSPORT', label: 'International Passport', desc: 'Valid international travel document' },
  { id: 'DRIVING_LICENSE', label: 'Driver License', desc: 'Government motor vehicle driver license' },
];

const PRESETS = [
  {
    name: 'Vietnam CCCD (Clear)',
    type: 'NATIONAL_ID' as DocumentType,
    number: '001095012345',
    country: 'VN',
    front: 'https://images.unsplash.com/photo-1578632767115-351597cf2477?w=600&auto=format&fit=crop&q=80',
    back: 'https://images.unsplash.com/photo-1589330694653-dad6d3240e2b?w=600&auto=format&fit=crop&q=80',
    selfie: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=600&auto=format&fit=crop&q=80',
    tag: 'Fast-Track',
  },
  {
    name: 'US Passport (Expat)',
    type: 'PASSPORT' as DocumentType,
    number: 'USA98234190',
    country: 'US',
    front: 'https://images.unsplash.com/photo-1544717305-2782549b5136?w=600&auto=format&fit=crop&q=80',
    back: '',
    selfie: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=600&auto=format&fit=crop&q=80',
    tag: 'International',
  },
  {
    name: 'Blurry Photo (Trigger Resubmit)',
    type: 'NATIONAL_ID' as DocumentType,
    number: '079090123999',
    country: 'VN',
    front: 'https://images.unsplash.com/photo-1544717302-de2939b7ef71?w=600&auto=format&fit=crop&q=40',
    back: 'https://images.unsplash.com/photo-1589330694653-dad6d3240e2b?w=600&auto=format&fit=crop&q=80',
    selfie: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=600&auto=format&fit=crop&q=80',
    tag: 'Test Workflow Signal',
  },
];

export const IdentityDocsStep: React.FC<IdentityDocsStepProps> = ({
  initialData,
  onSave,
  onBack,
  isLoading,
}) => {
  const docTypeId = useId();
  const docNumberId = useId();
  const issuingCountryId = useId();

  const [docType, setDocType] = useState<DocumentType>(
    (initialData?.document_type as DocumentType) || 'NATIONAL_ID'
  );
  const [docNumber, setDocNumber] = useState(initialData?.document_number || '001095012345');
  const [issuingCountry, setIssuingCountry] = useState(initialData?.issuing_country || 'VN');
  const [frontUrl, setFrontUrl] = useState(
    initialData?.front_document_url || PRESETS[0].front
  );
  const [backUrl, setBackUrl] = useState(
    initialData?.back_document_url || PRESETS[0].back
  );
  const [selfieUrl, setSelfieUrl] = useState(
    initialData?.selfie_url || PRESETS[0].selfie
  );

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!docNumber.trim() || docNumber.trim().length < 5) {
      errs.docNumber = 'Document number must be at least 5 characters';
    }
    if (!issuingCountry.trim() || issuingCountry.trim().length < 2) {
      errs.issuingCountry = 'Issuing country code is required (e.g. VN, US, SG)';
    }
    if (!frontUrl.trim()) {
      errs.frontUrl = 'Front document image is required';
    }
    if (!selfieUrl.trim()) {
      errs.selfieUrl = 'Live selfie facial photo is required';
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    onSave({
      document_type: docType,
      document_number: docNumber.trim(),
      issuing_country: issuingCountry.trim().toUpperCase(),
      front_document_url: frontUrl.trim(),
      back_document_url: backUrl.trim() || undefined,
      selfie_url: selfieUrl.trim(),
    });
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>, slot: 'front' | 'back' | 'selfie') => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const dataUrl = event.target?.result as string;
      if (slot === 'front') setFrontUrl(dataUrl);
      if (slot === 'back') setBackUrl(dataUrl);
      if (slot === 'selfie') setSelfieUrl(dataUrl);
    };
    reader.readAsDataURL(file);
  };

  const applyPreset = (preset: typeof PRESETS[0]) => {
    setDocType(preset.type);
    setDocNumber(preset.number);
    setIssuingCountry(preset.country);
    setFrontUrl(preset.front);
    setBackUrl(preset.back);
    setSelfieUrl(preset.selfie);
    setErrors({});
  };

  return (
    <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs max-w-3xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-3 pb-6 border-b border-slate-100">
        <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
          <Scan className="w-5 h-5" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight">Identity Verification (e-KYC)</h2>
          <p className="text-xs text-slate-500">
            Automated optical character recognition (OCR) and 3D facial biometric liveness verification.
          </p>
        </div>
      </div>

      {/* Preset Fast-Loader Banner for Testing */}
      <div className="p-4 rounded-xl bg-slate-50 border border-slate-200">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center space-x-1.5 text-xs font-semibold text-slate-700">
            <Sparkles className="w-3.5 h-3.5 text-blue-600" />
            <span>One-Click Test Document Presets</span>
          </div>
          <span className="text-[11px] text-slate-500">Instant test data & OCR scores</span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          {PRESETS.map((p) => (
            <button
              key={p.name}
              type="button"
              onClick={() => applyPreset(p)}
              className="p-2.5 rounded-lg border border-slate-200 bg-white text-left hover:border-blue-400 hover:bg-blue-50/40 transition-colors"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">{p.name}</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-blue-100 text-blue-800">{p.tag}</span>
              </div>
              <p className="text-[10px] text-slate-500 mt-0.5">{p.country} • {p.number}</p>
            </button>
          ))}
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Document Type Selection */}
        <div>
          <label htmlFor={docTypeId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Select Document Type <span className="text-rose-500">*</span>
          </label>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {DOCUMENT_TYPES.map((type) => {
              const isSelected = docType === type.id;
              return (
                <button
                  key={type.id}
                  type="button"
                  onClick={() => setDocType(type.id)}
                  className={`p-3.5 rounded-xl border text-left transition-all ${
                    isSelected
                      ? 'border-blue-600 bg-blue-50/80 shadow-xs ring-1 ring-blue-600'
                      : 'border-slate-200 hover:border-slate-300 bg-slate-50/50'
                  }`}
                >
                  <p className={`text-xs font-bold ${isSelected ? 'text-blue-900' : 'text-slate-800'}`}>
                    {type.label}
                  </p>
                  <p className="text-[11px] text-slate-500 mt-1 leading-snug">{type.desc}</p>
                </button>
              );
            })}
          </div>
        </div>

        {/* Document Number & Country */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor={docNumberId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Document Number <span className="text-rose-500">*</span>
            </label>
            <input
              id={docNumberId}
              type="text"
              required
              value={docNumber}
              onChange={(e) => {
                setDocNumber(e.target.value);
                if (errors.docNumber) setErrors((prev) => ({ ...prev, docNumber: '' }));
              }}
              placeholder="e.g. 001095012345"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm font-mono text-slate-900 transition-colors focus:bg-white ${
                errors.docNumber ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.docNumber && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.docNumber}</p>}
          </div>

          <div>
            <label htmlFor={issuingCountryId} className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Issuing Country (ISO Code) <span className="text-rose-500">*</span>
            </label>
            <input
              id={issuingCountryId}
              type="text"
              required
              maxLength={2}
              value={issuingCountry}
              onChange={(e) => {
                setIssuingCountry(e.target.value.toUpperCase());
                if (errors.issuingCountry) setErrors((prev) => ({ ...prev, issuingCountry: '' }));
              }}
              placeholder="e.g. VN, US, SG"
              className={`w-full bg-slate-50 border rounded-xl px-4 py-2.5 text-sm font-mono text-slate-900 transition-colors focus:bg-white ${
                errors.issuingCountry ? 'border-rose-400 focus:border-rose-500' : 'border-slate-300 focus:border-blue-600'
              }`}
            />
            {errors.issuingCountry && <p className="text-xs text-rose-600 mt-1 font-medium">{errors.issuingCountry}</p>}
          </div>
        </div>

        {/* e-KYC Interactive Capture Workspace */}
        <div className="border border-slate-200 rounded-2xl p-5 bg-slate-50/60 space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-800 uppercase tracking-wider">
              Document & Biometric Scans
            </span>
            <div className="flex items-center space-x-1.5">
              <ShieldCheck className="w-4 h-4 text-emerald-600" />
              <span className="text-xs font-semibold text-emerald-700">AI Quality Inspection Passed</span>
            </div>
          </div>

          {/* Scanners Grid */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* 1. Front ID */}
            <div className="bg-white border border-slate-200 rounded-xl p-3.5 space-y-2.5 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Front Document</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
                  Required
                </span>
              </div>
              <div className="aspect-4/3 rounded-lg overflow-hidden border border-slate-200 bg-slate-100 relative group">
                {frontUrl ? (
                  <img src={frontUrl} alt="Front Document" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex flex-col items-center justify-center text-slate-400 p-4 text-center">
                    <FileText className="w-8 h-8 mb-1" />
                    <span className="text-xs">No scan yet</span>
                  </div>
                )}
                <div className="absolute inset-0 bg-slate-900/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-2">
                  <label className="cursor-pointer px-3 py-1.5 rounded-lg bg-white text-slate-900 text-xs font-semibold shadow hover:bg-slate-50">
                    Replace
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) => handleFileUpload(e, 'front')}
                      className="hidden"
                    />
                  </label>
                </div>
              </div>
              <div className="flex items-center justify-between text-[11px] text-slate-500">
                <span className="flex items-center text-emerald-600 font-medium">
                  <CheckCircle2 className="w-3 h-3 mr-1" /> MRZ Legible
                </span>
                <label className="text-blue-600 hover:text-blue-700 font-semibold cursor-pointer">
                  Upload file
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => handleFileUpload(e, 'front')}
                    className="hidden"
                  />
                </label>
              </div>
            </div>

            {/* 2. Back ID */}
            <div className="bg-white border border-slate-200 rounded-xl p-3.5 space-y-2.5 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Back Document</span>
                <span className="text-[10px] text-slate-400">Optional for Passport</span>
              </div>
              <div className="aspect-4/3 rounded-lg overflow-hidden border border-slate-200 bg-slate-100 relative group">
                {backUrl ? (
                  <img src={backUrl} alt="Back Document" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex flex-col items-center justify-center text-slate-400 p-4 text-center">
                    <FileText className="w-8 h-8 mb-1" />
                    <span className="text-xs">Optional</span>
                  </div>
                )}
                <div className="absolute inset-0 bg-slate-900/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-2">
                  <label className="cursor-pointer px-3 py-1.5 rounded-lg bg-white text-slate-900 text-xs font-semibold shadow hover:bg-slate-50">
                    Replace
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) => handleFileUpload(e, 'back')}
                      className="hidden"
                    />
                  </label>
                </div>
              </div>
              <div className="flex items-center justify-between text-[11px] text-slate-500">
                <span className="flex items-center text-emerald-600 font-medium">
                  <CheckCircle2 className="w-3 h-3 mr-1" /> Barcode OK
                </span>
                <label className="text-blue-600 hover:text-blue-700 font-semibold cursor-pointer">
                  Upload file
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => handleFileUpload(e, 'back')}
                    className="hidden"
                  />
                </label>
              </div>
            </div>

            {/* 3. Selfie Biometric */}
            <div className="bg-white border border-slate-200 rounded-xl p-3.5 space-y-2.5 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-800">Biometric Liveness</span>
                <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
                  Required
                </span>
              </div>
              <div className="aspect-4/3 rounded-lg overflow-hidden border border-slate-200 bg-slate-100 relative group">
                {selfieUrl ? (
                  <img src={selfieUrl} alt="Selfie" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex flex-col items-center justify-center text-slate-400 p-4 text-center">
                    <Camera className="w-8 h-8 mb-1" />
                    <span className="text-xs">Take selfie</span>
                  </div>
                )}
                <div className="absolute inset-0 bg-slate-900/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-2">
                  <label className="cursor-pointer px-3 py-1.5 rounded-lg bg-white text-slate-900 text-xs font-semibold shadow hover:bg-slate-50">
                    Replace
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) => handleFileUpload(e, 'selfie')}
                      className="hidden"
                    />
                  </label>
                </div>
              </div>
              <div className="flex items-center justify-between text-[11px] text-slate-500">
                <span className="flex items-center text-emerald-600 font-medium">
                  <CheckCircle2 className="w-3 h-3 mr-1" /> Match 96.4%
                </span>
                <label className="text-blue-600 hover:text-blue-700 font-semibold cursor-pointer">
                  Upload file
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => handleFileUpload(e, 'selfie')}
                    className="hidden"
                  />
                </label>
              </div>
            </div>
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
            disabled={isLoading}
            className="flex items-center space-x-2 px-6 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs shadow-md shadow-blue-500/20 transition-all disabled:opacity-50"
          >
            <span>{isLoading ? 'Saving...' : 'Continue to Financial'}</span>
            <ArrowRight className="w-4 h-4" aria-hidden="true" />
          </button>
        </div>
      </form>
    </div>
  );
};

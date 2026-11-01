import React, { useEffect, useState } from 'react';
import { History, X, RefreshCw, Clock } from 'lucide-react';
import { onboardingApi } from '../../api/onboardingApi';
import { AuditLog } from '../../types/onboarding';

interface AuditLogsDrawerProps {
  applicationId: string;
  isOpen: boolean;
  onClose: () => void;
}

export const AuditLogsDrawer: React.FC<AuditLogsDrawerProps> = ({
  applicationId,
  isOpen,
  onClose,
}) => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchLogs = async () => {
    if (!applicationId) return;
    setIsLoading(true);
    try {
      const data = await onboardingApi.getAuditLogs(applicationId);
      setLogs(data || []);
    } catch (err) {
      console.error('Failed to load audit logs:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchLogs();
    }
  }, [isOpen, applicationId]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/40 backdrop-blur-xs animate-fadeIn">
      <div className="w-full max-w-md bg-white border-l border-slate-200 h-full flex flex-col shadow-2xl">
        {/* Header */}
        <div className="p-4 sm:p-5 border-b border-slate-200 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <History className="w-5 h-5 text-blue-600" />
            <h3 className="font-bold text-slate-900 text-base">Application Audit Trail</h3>
          </div>
          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={fetchLogs}
              disabled={isLoading}
              className="p-1.5 rounded-lg border border-slate-200 text-slate-500 hover:text-slate-900 hover:bg-slate-50 transition"
              title="Refresh logs"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-lg border border-slate-200 text-slate-500 hover:text-slate-900 hover:bg-slate-50 transition"
              title="Close drawer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Content list */}
        <div className="flex-1 overflow-y-auto p-4 sm:p-5 space-y-3">
          {isLoading && logs.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-xs">
              <RefreshCw className="w-5 h-5 animate-spin mx-auto mb-2 text-blue-600" />
              <span>Fetching audit timeline...</span>
            </div>
          ) : logs.length === 0 ? (
            <div className="py-12 text-center text-slate-400 text-xs">
              No audit logs recorded for this session yet.
            </div>
          ) : (
            logs.map((log) => (
              <div
                key={log.id}
                className="p-3.5 rounded-xl border border-slate-200/80 bg-slate-50 space-y-1.5 text-xs"
              >
                <div className="flex justify-between items-start">
                  <span className="font-bold font-mono text-blue-700">{log.event_type}</span>
                  <span className="text-[10px] text-slate-400 flex items-center">
                    <Clock className="w-3 h-3 mr-1" />
                    {new Date(log.created_at).toLocaleTimeString()}
                  </span>
                </div>

                {(log.from_status || log.to_status) && (
                  <div className="flex items-center space-x-1 text-[11px]">
                    <span className="text-slate-500">{log.from_status || 'INIT'}</span>
                    <span className="text-slate-400">→</span>
                    <span className="font-semibold text-slate-900">{log.to_status}</span>
                  </div>
                )}

                {log.reason && (
                  <p className="text-[11px] text-slate-600 italic">"{log.reason}"</p>
                )}

                {log.details && Object.keys(log.details).length > 0 && (
                  <pre className="mt-1 p-2 rounded-lg bg-slate-100 text-[10px] text-slate-700 font-mono overflow-x-auto">
                    {JSON.stringify(log.details, null, 2)}
                  </pre>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

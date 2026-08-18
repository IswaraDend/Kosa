import { useState } from 'react';
import { UploadCloud } from 'lucide-react';
import { api, ApiError } from '../lib/api';
import Modal from './Modal';

interface ImportRowError {
  row: number;
  message: string;
}

interface ImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  endpoint: string;
  projectId?: string;
  templateHint: string;
  onSuccess: () => void;
}

const ImportModal = ({ isOpen, onClose, title, endpoint, projectId, templateHint, onSuccess }: ImportModalProps) => {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const [rowErrors, setRowErrors] = useState<ImportRowError[]>([]);

  const reset = () => {
    setFile(null);
    setUploading(false);
    setSuccessMsg('');
    setErrorMsg('');
    setRowErrors([]);
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const handleUpload = async () => {
    if (!file) return;
    setUploading(true);
    setErrorMsg('');
    setRowErrors([]);
    setSuccessMsg('');
    try {
      const formData = new FormData();
      formData.append('file', file);
      if (projectId) formData.append('project_id', projectId);
      const res = await api.upload<{ imported_rows: number }>(endpoint, formData);
      setSuccessMsg(`${res.imported_rows} baris berhasil diimport`);
      setFile(null);
      onSuccess();
    } catch (err) {
      if (err instanceof ApiError && Array.isArray((err.body as { errors?: ImportRowError[] })?.errors)) {
        setRowErrors((err.body as { errors: ImportRowError[] }).errors);
      } else {
        setErrorMsg(err instanceof ApiError ? err.message : 'Gagal mengimport file');
      }
    } finally {
      setUploading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={title}
      footer={
        <>
          <button className="btn-secondary" onClick={handleClose}>
            Tutup
          </button>
          <button className="btn-primary" onClick={handleUpload} disabled={!file || uploading}>
            {uploading ? 'Mengimport...' : 'Import'}
          </button>
        </>
      }
    >
      <p style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12 }}>{templateHint}</p>

      <div className="form-group">
        <label>File Excel (.xlsx)</label>
        <input
          type="file"
          accept=".xlsx"
          className="form-input"
          onChange={(e) => {
            setFile(e.target.files?.[0] ?? null);
            setSuccessMsg('');
            setErrorMsg('');
            setRowErrors([]);
          }}
        />
      </div>

      {successMsg && (
        <div style={{ color: 'var(--success)', fontSize: 13, marginTop: 8 }}>{successMsg}</div>
      )}

      {errorMsg && <div style={{ color: 'var(--danger)', fontSize: 13, marginTop: 8 }}>{errorMsg}</div>}

      {rowErrors.length > 0 && (
        <div style={{ marginTop: 12 }}>
          <div style={{ color: 'var(--danger)', fontSize: 13, marginBottom: 6 }}>
            Import ditolak — perbaiki baris berikut lalu upload ulang:
          </div>
          <div style={{ maxHeight: 200, overflowY: 'auto', border: '1px solid var(--border-color)', borderRadius: 8 }}>
            <table>
              <thead>
                <tr>
                  <th>BARIS</th>
                  <th>PESAN</th>
                </tr>
              </thead>
              <tbody>
                {rowErrors.map((e, i) => (
                  <tr key={i}>
                    <td>{e.row}</td>
                    <td style={{ color: 'var(--danger)' }}>{e.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {!file && !successMsg && rowErrors.length === 0 && !errorMsg && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--text-muted)', fontSize: 12, marginTop: 4 }}>
          <UploadCloud size={14} /> Pilih file .xlsx untuk mulai import
        </div>
      )}
    </Modal>
  );
};

export default ImportModal;

import { X, Printer } from 'lucide-react';
import { formatCurrency } from '../lib/format';
import type { InvoiceRecord } from '../types';
import './InvoicePrintView.css';

interface InvoicePrintViewProps {
  invoice: InvoiceRecord;
  onClose: () => void;
}

const statusLabel: Record<string, string> = { unpaid: 'Belum Lunas', paid: 'Lunas', cancelled: 'Dibatalkan' };

const InvoicePrintView = ({ invoice, onClose }: InvoicePrintViewProps) => {
  return (
    <div className="invoice-print-overlay">
      <div className="invoice-print-toolbar">
        <button className="btn-secondary" onClick={onClose}>
          <X size={16} />
          Tutup
        </button>
        <button className="btn-primary" onClick={() => window.print()}>
          <Printer size={16} />
          Cetak
        </button>
      </div>

      <div className="invoice-print-area">
        <div className="invoice-print-header">
          <h2>INVOICE</h2>
          <div className="invoice-print-meta">
            <div>
              No. Invoice: <strong>{invoice.invoice_number}</strong>
            </div>
            <div>Tanggal: {new Date(invoice.created_at).toLocaleDateString('id-ID')}</div>
            <div>Status: {statusLabel[invoice.status] ?? invoice.status}</div>
          </div>
        </div>

        <div className="invoice-print-parties">
          <div>
            <div className="invoice-print-label">Ditagihkan kepada</div>
            <div>{invoice.customer?.name ?? '-'}</div>
            {invoice.customer?.address && <div>{invoice.customer.address}</div>}
            {invoice.customer?.phone && <div>{invoice.customer.phone}</div>}
          </div>
          <div>
            <div className="invoice-print-label">Dikirim dari Gudang</div>
            <div>{invoice.warehouse?.name ?? '-'}</div>
          </div>
        </div>

        <table className="invoice-print-table">
          <thead>
            <tr>
              <th>Produk</th>
              <th>Qty</th>
              <th>Harga Satuan</th>
              <th>Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {invoice.items.map((line) => (
              <tr key={line.id}>
                <td>{line.product?.name ?? `Produk #${line.product_id}`}</td>
                <td>
                  {line.quantity} {line.product?.unit ?? ''}
                </td>
                <td>{formatCurrency(line.unit_price)}</td>
                <td>{formatCurrency(line.unit_price * line.quantity)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={3}>Total</td>
              <td>{formatCurrency(invoice.subtotal)}</td>
            </tr>
          </tfoot>
        </table>

        {invoice.note && (
          <div className="invoice-print-note">
            <div className="invoice-print-label">Catatan</div>
            <div>{invoice.note}</div>
          </div>
        )}
      </div>
    </div>
  );
};

export default InvoicePrintView;

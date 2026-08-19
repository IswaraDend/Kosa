import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Mail, Lock, Eye, Server, Warehouse, ArrowRightLeft, ArrowRight } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import './Login.css';

const Login = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg('');

    try {
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      const response = await fetch(`${apiUrl}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (response.ok) {
        login(data.token, data.user);

        // Routing berdasarkan role user
        if (data.user.is_super_admin) {
          navigate('/super-admin');
        } else if (data.user.is_admin) {
          navigate('/admin');
        } else if (data.user.is_member) {
          navigate('/member');
        } else {
          navigate('/login');
        }
      } else {
        setErrorMsg(data.error || 'Login gagal');
      }
    } catch {
      setErrorMsg('Gagal terhubung ke server backend');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-left">
        <svg className="login-rings" viewBox="0 0 520 520" aria-hidden="true">
          <circle cx="260" cy="260" r="230" />
          <circle cx="260" cy="260" r="185" />
          <circle cx="260" cy="260" r="140" />
          <circle cx="260" cy="260" r="95" />
          <circle cx="260" cy="260" r="50" />
        </svg>

        <div className="brand">
          <span className="brand-mark">कोष</span>
          <span className="brand-word">Kośa</span>
        </div>

        <div className="login-hero">
          <h1>Satu dashboard untuk stok semua project-mu.</h1>
          <p>Masuk untuk memantau stok, gudang, dan transaksi secara real-time dari mana saja.</p>
        </div>

        <div className="features">
          <div className="feature-item">
            <div className="feature-icon"><Server size={20} /></div>
            <span>Kelola stok lintas puluhan project dalam satu dashboard</span>
          </div>
          <div className="feature-item">
            <div className="feature-icon"><Warehouse size={20} /></div>
            <span>Pantau banyak gudang per project secara real-time</span>
          </div>
          <div className="feature-item">
            <div className="feature-icon"><ArrowRightLeft size={20} /></div>
            <span>Transfer stok antar project tanpa kehilangan histori</span>
          </div>
        </div>
      </div>

      <div className="login-right">
        <div className="login-box">
          <div className="login-subtitle">MASUK KE AKUN</div>
          <h2>Selamat datang kembali</h2>
          <div className="login-register-text">
            Belum punya akun? <a href="#">Buat akun baru</a>
          </div>

          <form onSubmit={handleLogin}>
            {errorMsg && (
              <div style={{ color: 'var(--danger)', marginBottom: '16px', fontSize: '14px', textAlign: 'center', backgroundColor: 'rgba(231, 76, 60, 0.1)', padding: '10px', borderRadius: '8px' }}>
                {errorMsg}
              </div>
            )}
            <div className="form-group">
              <label>Email</label>
              <div className="input-wrapper">
                <Mail size={18} className="input-icon" />
                <input 
                  type="email" 
                  className="input-field" 
                  placeholder="nama@perusahaan.com (atau iswaradend@gmail.com)"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
            </div>

            <div className="form-group">
              <label>
                Kata sandi
                <a href="#">Lupa sandi?</a>
              </label>
              <div className="input-wrapper">
                <Lock size={18} className="input-icon" />
                <input 
                  type="password" 
                  className="input-field" 
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
                <Eye size={18} className="input-icon-right" />
              </div>
            </div>

            <label className="remember-me">
              <input type="checkbox" defaultChecked />
              Ingat saya di perangkat ini
            </label>

            <button type="submit" className="btn-primary" disabled={loading} style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
              {loading ? 'Memproses...' : 'Masuk'} <ArrowRight size={16} style={{ marginLeft: 8 }} />
            </button>
          </form>

          <p className="terms-text">
            Dengan masuk, kamu menyetujui Ketentuan Layanan dan Kebijakan Privasi Kosa.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;

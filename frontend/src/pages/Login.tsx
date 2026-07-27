import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { GoogleLogin } from '@react-oauth/google';
import { Mail, Lock, Eye, Radio, Server, Warehouse, ArrowRightLeft, ArrowRight } from 'lucide-react';
import './Login.css';

const Login = () => {
  const navigate = useNavigate();
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
        // Simpan token dan data user
        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
        
        // Routing berdasarkan RBAC baru
        if (data.user.is_super_admin) {
          navigate('/super-admin');
        } else {
          // Sementara arahkan ke Admin, nanti bisa dicek lagi role spesifiknya
          navigate('/admin');
        }
      } else {
        setErrorMsg(data.error || 'Login gagal');
      }
    } catch (err) {
      setErrorMsg('Gagal terhubung ke server backend');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-left">
        <div className="brand">
          <div className="brand-icon">
            <Radio size={20} />
          </div>
          Stockpulse
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

            <button type="submit" className="btn-primary" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
              Masuk <ArrowRight size={16} style={{ marginLeft: 8 }} />
            </button>
          </form>

          <div className="divider">ATAU</div>
          
          <div className="sso-btn-wrapper">
            <GoogleLogin
              onSuccess={credentialResponse => {
                console.log(credentialResponse);
                navigate('/super-admin'); 
              }}
              onError={() => {
                console.log('Login Failed');
              }}
              theme="filled_black"
              shape="rectangular"
              text="signin_with"
              width="400px"
            />
          </div>
          
          <p className="terms-text">
            Dengan masuk, kamu menyetujui Ketentuan Layanan dan Kebijakan Privasi Stockpulse.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;

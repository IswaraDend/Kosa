import React from 'react';
import ReactDOM from 'react-dom/client';
import { GoogleOAuthProvider } from '@react-oauth/google';
import App from './App.tsx';
import { AuthProvider } from './context/AuthContext';
import { ProjectProvider } from './context/ProjectContext';
import './index.css';

// TODO: Ganti dengan Google Client ID Anda dari Google Cloud Console
const GOOGLE_CLIENT_ID = "GANTI_DENGAN_CLIENT_ID_ANDA.apps.googleusercontent.com";

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
      <AuthProvider>
        <ProjectProvider>
          <App />
        </ProjectProvider>
      </AuthProvider>
    </GoogleOAuthProvider>
  </React.StrictMode>,
);

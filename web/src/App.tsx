import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Toaster } from 'sonner';

import { Layout } from '@/components/Layout';
import { Dashboard } from '@/pages/Dashboard';
import { RunbooksPage } from '@/pages/RunbooksPage';
import { BuilderPage } from '@/pages/BuilderPage';
import { ExecutionsPage } from '@/pages/ExecutionsPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { LoginPage } from '@/pages/LoginPage';
import { AuthProvider } from '@/contexts/AuthContext';
import { ThemeProvider } from '@/contexts/ThemeContext';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthProvider>
          <Router>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route
                path="/*"
                element={
                  <Layout>
                    <Routes>
                      <Route path="/" element={<Navigate to="/dashboard" replace />} />
                      <Route path="/dashboard" element={<Dashboard />} />
                      <Route path="/runbooks" element={<RunbooksPage />} />
                      <Route path="/runbooks/:id/edit" element={<BuilderPage />} />
                      <Route path="/runbooks/new" element={<BuilderPage />} />
                      <Route path="/executions" element={<ExecutionsPage />} />
                      <Route path="/settings" element={<SettingsPage />} />
                    </Routes>
                  </Layout>
                }
              />
            </Routes>
            <Toaster position="top-right" />
          </Router>
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;

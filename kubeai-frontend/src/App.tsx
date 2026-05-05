import { Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import LoginPage from './pages/Login';
import RegisterPage from './pages/Register';
import DashboardPage from './pages/Dashboard';
import ModelListPage from './pages/models/ModelList';
import ModelDetailPage from './pages/models/ModelDetail';
import ModelCreatePage from './pages/models/ModelCreate';
import VersionCreatePage from './pages/models/VersionCreate';
import InferenceExecutePage from './pages/inference/InferenceExecute';
import JobListPage from './pages/jobs/JobList';
import TrainingJobCreatePage from './pages/jobs/TrainingJobCreate';
import InferenceJobCreatePage from './pages/jobs/InferenceJobCreate';
import HealthCheckPage from './pages/health/HealthCheck';
import ProtectedRoute from './components/ProtectedRoute';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <MainLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="models" element={<ModelListPage />} />
        <Route path="models/create" element={<ModelCreatePage />} />
        <Route path="models/:name" element={<ModelDetailPage />} />
        <Route path="models/:name/versions/create" element={<VersionCreatePage />} />
        <Route path="inference/execute" element={<InferenceExecutePage />} />
        <Route path="jobs" element={<JobListPage />} />
        <Route path="jobs/training/create" element={<TrainingJobCreatePage />} />
        <Route path="jobs/inference/create" element={<InferenceJobCreatePage />} />
        <Route path="health" element={<HealthCheckPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
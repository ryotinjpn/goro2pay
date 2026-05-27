import { Route, Routes } from 'react-router-dom';
import LandingMock from './screens/LandingMock';
import MainMock from './screens/MainMock';
import CompleteMock from './screens/CompleteMock';

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingMock />} />
      <Route path="/main" element={<MainMock />} />
      <Route path="/main/complete" element={<CompleteMock />} />
    </Routes>
  );
}

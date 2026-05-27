import { Route, Routes } from 'react-router-dom';
import LandingMock from './screens/LandingMock';
import MainMock from './screens/MainMock';
import CompleteMock from './screens/CompleteMock';
import SignupMock from './screens/SignupMock';
import LoginMock from './screens/LoginMock';
import BudgetSetupMock from './screens/BudgetSetupMock';
import NotFoundMock from './screens/NotFoundMock';
import ErrorMock from './screens/ErrorMock';
import LogoutModalMock from './screens/LogoutModalMock';
import SessionExpiredModalMock from './screens/SessionExpiredModalMock';
import InsufficientBalanceModalMock from './screens/InsufficientBalanceModalMock';
import RaiseBudgetModalMock from './screens/RaiseBudgetModalMock';

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingMock />} />
      <Route path="/main" element={<MainMock />} />
      <Route path="/main/complete" element={<CompleteMock />} />
      <Route path="/signup" element={<SignupMock />} />
      <Route path="/login" element={<LoginMock />} />
      <Route path="/budget" element={<BudgetSetupMock />} />
      <Route path="/error" element={<ErrorMock />} />
      <Route path="/logout-modal" element={<LogoutModalMock />} />
      <Route path="/session-expired" element={<SessionExpiredModalMock />} />
      <Route path="/insufficient" element={<InsufficientBalanceModalMock />} />
      <Route path="/raise-budget" element={<RaiseBudgetModalMock />} />
      <Route path="/not-found" element={<NotFoundMock />} />
      <Route path="*" element={<NotFoundMock />} />
    </Routes>
  );
}

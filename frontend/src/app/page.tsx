"use client";

import { useState } from "react";
import { useAuth } from "../context/AuthContext";

export default function Home() {
  const { user, login, logout } = useAuth();
  const [showLogin, setShowLogin] = useState(true);
  const [message, setMessage] = useState("");
  const [isError, setIsError] = useState(false);

  const showMessage = (msg: string, error: boolean = false) => {
    setMessage(msg);
    setIsError(error);
    setTimeout(() => {
      setMessage("");
    }, 3000);
  };

  const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    const avatarFile = formData.get("avatar") as File;

    let avatarPath = "";
    if (avatarFile && avatarFile.size > 0) {
      try {
        const uploadFormData = new FormData();
        uploadFormData.append("image", avatarFile);
        const uploadRes = await fetch(`http://localhost:8080/api/upload`, {
          method: "POST",
          body: uploadFormData,
          credentials: "include",
        });

        if (uploadRes.ok) {
          const uploadData = await uploadRes.json();
          avatarPath = uploadData.path;
        } else {
          const error = await uploadRes.text();
          showMessage(`Avatar upload failed: ${error}`, true);
          return;
        }
      } catch (err: any) {
        showMessage(`Avatar upload failed: ${err.message}`, true);
        return;
      }
    }

    const registrationData = { ...data, avatar: avatarPath };

    try {
      const res = await fetch(`http://localhost:8080/api/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(registrationData),
        credentials: "include",
      });

      if (res.ok) {
        showMessage("Registration successful! Please log in.");
        setShowLogin(true);
        form.reset();
      } else {
        const error = await res.text();
        showMessage(`Registration failed: ${error}`, true);
      }
    } catch (err: any) {
      showMessage(`Registration failed: ${err.message}`, true);
    }
  };

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());

    try {
      const res = await fetch(`http://localhost:8080/api/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
        credentials: "include",
      });

      if (res.ok) {
        const resData = await res.json();
        login(resData);
        showMessage("Login successful!");
      } else {
        const error = await res.text();
        showMessage(`Login failed: ${error}`, true);
      }
    } catch (err: any) {
      showMessage(`Login failed: ${err.message}`, true);
    }
  };

  return (
    <main>
      <h1>Social Dilemma</h1>
      <div className="container">
        {user ? (
          <div id="content">
            <h2>Welcome to your Dashboard, {user.first_name}!</h2>
            <p>Use the navigation bar to explore the application.</p>
          </div>
        ) : (
          <div className="auth-container">
            {showLogin ? (
              <div id="login-form" className="form-container">
                <h2>Login</h2>
                <form id="login" onSubmit={handleLogin}>
                  <input
                    type="email"
                    name="email"
                    placeholder="Email"
                    required
                  />
                  <input
                    type="password"
                    name="password"
                    placeholder="Password"
                    required
                  />
                  <button type="submit">Login</button>
                </form>
                <p>
                  Don't have an account?{" "}
                  <button
                    onClick={() => setShowLogin(false)}
                    className="toggle-btn"
                  >
                    Register
                  </button>
                </p>
              </div>
            ) : (
              <div id="register-form" className="form-container">
                <h2>Register</h2>
                <form id="register" onSubmit={handleRegister}>
                  <input
                    type="text"
                    name="first_name"
                    placeholder="First Name"
                    required
                  />
                  <input
                    type="text"
                    name="last_name"
                    placeholder="Last Name"
                    required
                  />
                  <input
                    type="email"
                    name="email"
                    placeholder="Email"
                    required
                  />
                  <input
                    type="password"
                    name="password"
                    placeholder="Password"
                    required
                  />
                  <input
                    type="date"
                    name="date_of_birth"
                    placeholder="Date of Birth"
                    required
                  />
                  <input type="file" name="avatar" accept="image/*" />
                  <button type="submit">Register</button>
                </form>
                <p>
                  Already have an account?{" "}
                  <button
                    onClick={() => setShowLogin(true)}
                    className="toggle-btn"
                  >
                    Login
                  </button>
                </p>
              </div>
            )}
          </div>
        )}
        {message && (
          <div
            id="message"
            className={isError ? "error" : "success"}
          >
            {message}
          </div>
        )}
      </div>
    </main>
  );
}

import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="container" style={{ textAlign: 'center', marginTop: '5rem' }}>
      <div className="card">
        <h1 style={{ color: 'var(--error-color)', fontSize: '5rem', marginBottom: '1rem' }}>404</h1>
        <h2>Page Not Found</h2>
        <p style={{ color: 'var(--text-muted)', marginTop: '0.5rem', marginBottom: '2rem' }}>
          Sorry, the page you are looking for does not exist or has been moved.
        </p>
        <Link href="/" className="btn">
          Go back to Home
        </Link>
      </div>
    </div>
  );
}

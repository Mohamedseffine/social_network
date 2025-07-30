'use client'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import RenderHomePage from "../components/home"
import axios from 'axios'
import styles from './globals.module.css' // Create this if needed

export default function HomePage() {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(true)

  const checkSession = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/check-session', {
        withCredentials: true,
      })
      console.log('Session valid:', response.data)
      setIsLoading(false)
    } catch (error) {
      console.error('Failed to check session:', error)
      router.push('/login')
    }
  }

  useEffect(() => {
    checkSession()
  }, [])

  if (isLoading) {
    return (
      <div className={styles.loadingContainer}>
        <div className={styles.loadingSpinner}></div>
        <p>Verifying session...</p>
      </div>
    )
  }

  return (
    <main>
      <RenderHomePage />
    </main>
  )
}
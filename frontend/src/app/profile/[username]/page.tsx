'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import ProfilePage from '@/app/ProfileComponent'; // Extracted your big component into a reusable one

export default function PublicProfilePage() {
  const params = useParams();
  const username = params?.username as string | undefined;

  return <ProfilePage username={username} />;
}

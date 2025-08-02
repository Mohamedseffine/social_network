import React, { useState, useEffect } from 'react';
import axios from 'axios';
import styles from './css/FollowButton.module.css';

interface FollowButtonProps {
  id: number
}

type FollowStatus = 'none' | 'pending' | 'accepted' | 'declined';

const FollowButton: React.FC<FollowButtonProps> = ({ id }) => {
  console.log("The following id id: ", id);

  const [status, setStatus] = useState<FollowStatus>('none');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

useEffect(() => {
  const fetchFollowStatus = async () => {
    try {
      const res = await axios.get(
        'http://localhost:8080/api/follow/status',
        {
          params: { target_id: id },
          withCredentials: true,
        }
      );
      console.log("following status: ", res.data);
      
      setStatus(res.data.status as FollowStatus);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  fetchFollowStatus();
}, [id]);


  const handleFollowRequest = async () => {
    setIsLoading(true);
    setError(null);

    try {
      await axios.post(
        'http://localhost:8080/api/follow',
        { target_id: id },
        { withCredentials: true }
      );
      setStatus('pending');
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const renderButtonText = () => {
    if (isLoading) return '...';
    switch (status) {
      case 'none':
        return 'Follow';
      case 'pending':
        return 'Request Sent';
      case 'accepted':
        return 'Following';
      case 'declined':
        return 'Declined';
      default:
        return 'follow';
    }
  };

  const isButtonDisabled = () =>
    isLoading || status === 'pending' || status === 'accepted';

  return (
    <div>
      {status !== 'accepted' && (
        <button
          className={styles.followButton}
          onClick={handleFollowRequest}
          disabled={isButtonDisabled()}
        >
          {renderButtonText()}
        </button>
      )}
      {error && <p className={styles.error}>{error}</p>}
    </div>

  );
};

export default FollowButton;

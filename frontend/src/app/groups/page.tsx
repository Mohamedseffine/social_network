"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { API_BASE_URL } from "../../utils/api";
import { useAuth } from "@/context/AuthContext";

const GroupsPage = () => {
  const [groups, setGroups] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchTerm, setSearchTerm] = useState("")
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const router = useRouter();
  const {user} = useAuth()

  const fetchGroups = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/groups`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setGroups(data);
      } else {
        const errorText = await res.text();
        setError(`Failed to fetch groups: ${errorText}`);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (searchTerm.trim() === "") {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    const delayDebounceFn = setTimeout(async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/search_groups?q=${searchTerm}`, {
          credentials: "include",
        });
        if (res.ok) {
          const data = await res.json();
          setSearchResults(data.groups?data.groups:[]);
        } else {
          console.error("Failed to search users");
          if (res.status == 401){
                      router.push("/")

          }
        }
      } catch (err) {
        console.error("User search error:", err);
      } finally {
        setIsSearching(false);
      }
    }, 500); // 500ms delay
    return () => clearTimeout(delayDebounceFn);
  }, [searchTerm]);

  useEffect(() => {
    fetchGroups();
  }, []);

  const handleCreateGroup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const title = formData.get("title");
    const description = formData.get("description");

    try {
      const res = await fetch(`${API_BASE_URL}/groups`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description }),
        credentials: "include",
      });

      if (res.ok) {
        form.reset();
        fetchGroups(); // Refresh groups
      } else {
        const errorText = await res.text();
        setError(`Failed to create group: ${errorText}`);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
   user != null &&( <div className="groups-container">
      <h1>Groups</h1>
      <div className="create-group">
        <h2>Create a Group</h2>
        <form onSubmit={handleCreateGroup}>
          <input type="text" name="title" placeholder="Group Title" required />
          <textarea
            name="description"
            placeholder="Group Description"
            required
          ></textarea>
          <button type="submit">Create Group</button>
        </form>
      </div>
      <div>
        <h3>Search For A Group</h3>
              <input
                type="text"
                placeholder="Search by name or nickname..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              {isSearching && <div>Searching...</div>}
              <div className="search-results">
                { searchResults.length>0 && searchResults?.map((group) => (
                  <div key={group.id} className="search-result-item">
                    <span><strong>title:</strong>{group.title+"  "}</span>
                    <span><strong>discreption:</strong>{group.description}</span>
                    <Link href={`/groups/${group.id}`} > see group</Link>
                  </div>
                ))}
                {searchResults.length===0 && !isSearching  && searchTerm&&(
                  <div>can't find any results</div>
                )}
                </div>
      </div>
      <div className="groups-list">
        <h2>All Groups</h2>
        {groups && groups.length > 0 ? (
          groups.map((group) => (
            <div key={group.id} className="group-item">
              <Link href={`/groups/${group.id}`}>
                <h3>{group.title}</h3>
              </Link>
              <p>{group.description}</p>
            </div>
          ))
        ) : (
          <p>No groups found.</p>
        )}
      </div>
    </div>)
  );
};

export default GroupsPage;

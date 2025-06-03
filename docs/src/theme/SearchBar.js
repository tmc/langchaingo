import React, { useState, useEffect, useRef } from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import './SearchBar.css';

export default function SearchBar() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [searchIndex, setSearchIndex] = useState([]);
  const searchRef = useRef(null);
  const { siteConfig } = useDocusaurusContext();

  // Load search index
  useEffect(() => {
    const loadSearchIndex = async () => {
      try {
        const response = await fetch(`${siteConfig.baseUrl}search-index.json`);
        if (response.ok) {
          const index = await response.json();
          setSearchIndex(index);
        }
      } catch (error) {
        console.warn('Search index not available:', error);
        // Fallback to basic page search
        setSearchIndex([
          { title: 'Getting Started', url: '/docs/', content: 'LangChain Go documentation' },
          { title: 'Agents', url: '/docs/modules/agents/', content: 'Agent functionality' },
          { title: 'Chains', url: '/docs/modules/chains/', content: 'Chain operations' },
          { title: 'Models', url: '/docs/modules/model_io/models/', content: 'Language models' },
          { title: 'OpenAI', url: '/docs/modules/model_io/models/llms/Integrations/openai', content: 'OpenAI integration' },
          { title: 'Mistral', url: '/docs/modules/model_io/models/llms/Integrations/mistral', content: 'Mistral AI integration' },
          { title: 'Vector Stores', url: '/docs/modules/data_connection/vector_stores/', content: 'Vector database storage' },
          { title: 'Text Splitters', url: '/docs/modules/data_connection/text_splitters/', content: 'Document text splitting' },
          { title: 'Prompts', url: '/docs/modules/model_io/prompts/', content: 'Prompt templates and management' },
          { title: 'Memory', url: '/docs/modules/memory/', content: 'Conversation memory management' },
        ]);
      }
    };
    loadSearchIndex();
  }, [siteConfig.baseUrl]);

  // Handle search
  useEffect(() => {
    if (query.length < 2) {
      setResults([]);
      return;
    }

    const searchResults = searchIndex
      .filter(item => 
        item.title.toLowerCase().includes(query.toLowerCase()) ||
        item.content.toLowerCase().includes(query.toLowerCase())
      )
      .slice(0, 8)
      .map(item => ({
        ...item,
        highlight: highlightMatch(item.title, query) || highlightMatch(item.content, query)
      }));

    setResults(searchResults);
  }, [query, searchIndex]);

  const highlightMatch = (text, query) => {
    const index = text.toLowerCase().indexOf(query.toLowerCase());
    if (index === -1) return null;
    
    const start = Math.max(0, index - 20);
    const end = Math.min(text.length, index + query.length + 20);
    const snippet = text.slice(start, end);
    
    return snippet.replace(
      new RegExp(`(${query})`, 'gi'),
      '<mark>$1</mark>'
    );
  };

  const handleInputChange = (e) => {
    setQuery(e.target.value);
    setIsOpen(true);
  };

  const handleResultClick = (url) => {
    window.location.href = url;
    setQuery('');
    setIsOpen(false);
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
      setQuery('');
    }
  };

  // Close search when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (searchRef.current && !searchRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="navbar__search" ref={searchRef}>
      <input
        className="navbar__search-input"
        placeholder="Search docs..."
        value={query}
        onChange={handleInputChange}
        onFocus={() => setIsOpen(true)}
        onKeyDown={handleKeyDown}
        aria-label="Search"
      />
      
      {isOpen && results.length > 0 && (
        <div className="search-results">
          {results.map((result, index) => (
            <div
              key={index}
              className="search-result-item"
              onClick={() => handleResultClick(result.url)}
            >
              <div className="search-result-title">{result.title}</div>
              <div 
                className="search-result-content"
                dangerouslySetInnerHTML={{ __html: result.highlight || result.content }}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
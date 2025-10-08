import React, { useState, useEffect, useRef } from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import { useHistory } from '@docusaurus/router';
import './SearchBar.css';

export default function SearchBar() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [searchIndex, setSearchIndex] = useState([]);
  const searchRef = useRef(null);
  const context = useDocusaurusContext();
  const siteConfig = context?.siteConfig || {};
  const history = useHistory();

  // Load search index
  useEffect(() => {
    const loadSearchIndex = async () => {
      try {
        const response = await fetch(`${siteConfig.baseUrl || '/'}search-index.json`);
        if (response.ok) {
          const index = await response.json();
          setSearchIndex(index);
        }
      } catch (error) {
        console.warn('Search index not available:', error);
        // Fallback to basic page search
        setSearchIndex([
          { title: 'Getting Started', url: '/langchaingo/docs/', content: 'LangChainGo documentation' },
          { title: 'Tutorials', url: '/langchaingo/docs/tutorials/', content: 'Step-by-step guides to build complete applications' },
          { title: 'Building a Simple Chat Application', url: '/langchaingo/docs/tutorials/basic-chat-app', content: 'Learn the basics with conversation memory' },
          { title: 'How-to Guides', url: '/langchaingo/docs/how-to/', content: 'Practical solutions for specific problems' },
          { title: 'Configure LLM Providers', url: '/langchaingo/docs/how-to/configure-llm-providers', content: 'How to configure different LLM providers' },
          { title: 'Concepts', url: '/langchaingo/docs/concepts/', content: 'Core concepts and architecture' },
          { title: 'LangChainGo Architecture', url: '/langchaingo/docs/concepts/architecture', content: 'Architecture and design principles' },
          { title: 'Agents', url: '/langchaingo/docs/modules/agents/', content: 'Agent functionality' },
          { title: 'Chains', url: '/langchaingo/docs/modules/chains/', content: 'Chain operations' },
          { title: 'Models', url: '/langchaingo/docs/modules/model_io/models/', content: 'Language models' },
          { title: 'OpenAI', url: '/langchaingo/docs/modules/model_io/models/llms/Integrations/openai', content: 'OpenAI integration' },
          { title: 'Mistral', url: '/langchaingo/docs/modules/model_io/models/llms/Integrations/mistral', content: 'Mistral AI integration' },
          { title: 'Vector Stores', url: '/langchaingo/docs/modules/data_connection/vector_stores/', content: 'Vector database storage' },
          { title: 'PGVector', url: '/langchaingo/docs/modules/data_connection/vector_stores/pgvector', content: 'PostgreSQL vector storage' },
          { title: 'Text Splitters', url: '/langchaingo/docs/modules/data_connection/text_splitters/', content: 'Document text splitting' },
          { title: 'Prompts', url: '/langchaingo/docs/modules/model_io/prompts/', content: 'Prompt templates and management' },
          { title: 'Memory', url: '/langchaingo/docs/modules/memory/', content: 'Conversation memory management' },
          { title: 'API Reference', url: 'https://pkg.go.dev/github.com/tmc/langchaingo', content: 'Complete API documentation', external: true },
        ]);
      }
    };
    loadSearchIndex();
  }, [siteConfig?.baseUrl]);

  // Handle search
  useEffect(() => {
    if (query.length < 2) {
      setResults([]);
      return;
    }

    const searchResults = searchIndex
      .filter(item => {
        const queryLower = query.toLowerCase();
        
        // Search in title
        if (item.title.toLowerCase().includes(queryLower)) {
          return true;
        }
        
        // Search in content
        if (item.content && item.content.toLowerCase().includes(queryLower)) {
          return true;
        }
        
        // Search in package.title combination (e.g., "llms.Model")
        if (item.package && (item.package + '.' + item.title).toLowerCase().includes(queryLower)) {
          return true;
        }
        
        // Search in keywords array
        if (item.keywords && item.keywords.some(keyword => 
          keyword.toLowerCase().includes(queryLower)
        )) {
          return true;
        }
        
        // Search in signature
        if (item.signature && item.signature.toLowerCase().includes(queryLower)) {
          return true;
        }
        
        return false;
      })
      .slice(0, 8)
      .map(item => ({
        ...item,
        highlight: highlightMatch(item.title, query) || 
                   highlightMatch(item.content, query) ||
                   (item.package && highlightMatch(item.package + '.' + item.title, query)) ||
                   (item.signature && highlightMatch(item.signature, query))
      }));

    setResults(searchResults);
  }, [query, searchIndex]);

  const highlightMatch = (text, searchQuery) => {
    if (!text || !searchQuery) return null;
    
    const index = text.toLowerCase().indexOf(searchQuery.toLowerCase());
    if (index === -1) return null;
    
    const start = Math.max(0, index - 20);
    const end = Math.min(text.length, index + searchQuery.length + 20);
    const snippet = text.slice(start, end);
    
    return snippet.replace(
      new RegExp(`(${searchQuery})`, 'gi'),
      '<mark>$1</mark>'
    );
  };

  const handleInputChange = (e) => {
    setQuery(e.target.value);
    setIsOpen(true);
  };

  const handleResultClick = (url) => {
    if (url.startsWith('http')) {
      window.open(url, '_blank');
    } else {
      history.push(url);
    }
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
        onFocus={() => query.length >= 2 && setIsOpen(true)}
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
              <div className="search-result-title">
                {result.package && <span className="search-result-package">{result.package}.</span>}
                {result.title}
                {result.external && <span className="search-result-external">↗</span>}
              </div>
              <div 
                className="search-result-content"
                dangerouslySetInnerHTML={{ __html: result.highlight || result.content }}
              />
            </div>
          ))}
        </div>
      )}
      
      {isOpen && query.length >= 2 && results.length === 0 && (
        <div className="search-results">
          <div className="search-no-results">
            No results found for "{query}"
          </div>
        </div>
      )}
    </div>
  );
}
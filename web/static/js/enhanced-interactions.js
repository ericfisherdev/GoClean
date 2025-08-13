/* Enhanced GoClean Report Interactions */

class EnhancedReportManager {
    constructor() {
        this.initializeComponents();
        this.setupEventListeners();
        this.startPerformanceMonitoring();
    }

    initializeComponents() {
        this.themeManager = new ThemeManager();
        this.searchManager = new EnhancedSearchManager();
        this.animationManager = new AnimationManager();
        this.accessibilityManager = new AccessibilityManager();
        this.performanceManager = new PerformanceManager();
        this.exportManager = new ExportManager();
    }

    setupEventListeners() {
        // Global keyboard shortcuts
        document.addEventListener('keydown', this.handleGlobalKeydown.bind(this));
        
        // Window resize handling
        window.addEventListener('resize', this.debounce(this.handleResize.bind(this), 250));
        
        // Visibility change handling (for auto-refresh optimization)
        document.addEventListener('visibilitychange', this.handleVisibilityChange.bind(this));
    }

    handleGlobalKeydown(event) {
        // Global shortcuts
        if (event.ctrlKey || event.metaKey) {
            switch (event.key.toLowerCase()) {
                case 'k':
                    event.preventDefault();
                    this.searchManager.focusSearch();
                    break;
                case 't':
                    event.preventDefault();
                    this.themeManager.toggleTheme();
                    break;
                case 's':
                    event.preventDefault();
                    this.exportManager.showExportDialog();
                    break;
            }
        }
        
        if (event.key === 'Escape') {
            this.searchManager.clearSearch();
            this.closeAllModals();
        }
    }

    handleResize() {
        this.searchManager.handleResize();
        this.animationManager.handleResize();
    }

    handleVisibilityChange() {
        if (document.hidden) {
            this.performanceManager.pauseAnimations();
        } else {
            this.performanceManager.resumeAnimations();
        }
    }

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    startPerformanceMonitoring() {
        this.performanceManager.startMonitoring();
    }

    closeAllModals() {
        // Close any open modals or dropdowns
        const modals = document.querySelectorAll('.modal.show');
        modals.forEach(modal => {
            const bsModal = bootstrap.Modal.getInstance(modal);
            if (bsModal) bsModal.hide();
        });
    }
}

class ThemeManager {
    constructor() {
        this.currentTheme = localStorage.getItem('goclean-theme') || 'auto';
        this.applyTheme();
        this.setupThemeToggle();
    }

    setupThemeToggle() {
        // Create theme toggle button if it doesn't exist
        if (!document.querySelector('.theme-toggle')) {
            this.createThemeToggle();
        }
    }

    createThemeToggle() {
        const navbar = document.querySelector('.navbar-nav');
        if (navbar) {
            const themeToggle = document.createElement('div');
            themeToggle.className = 'nav-item dropdown';
            themeToggle.innerHTML = `
                <a class="nav-link dropdown-toggle theme-toggle" href="#" role="button" data-bs-toggle="dropdown">
                    <i class="bi bi-palette"></i>
                </a>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li><a class="dropdown-item" href="#" data-theme="light"><i class="bi bi-sun-fill"></i> Light</a></li>
                    <li><a class="dropdown-item" href="#" data-theme="dark"><i class="bi bi-moon-fill"></i> Dark</a></li>
                    <li><a class="dropdown-item" href="#" data-theme="auto"><i class="bi bi-circle-half"></i> Auto</a></li>
                </ul>
            `;
            navbar.appendChild(themeToggle);

            // Add event listeners
            themeToggle.querySelectorAll('[data-theme]').forEach(link => {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.setTheme(e.target.closest('[data-theme]').dataset.theme);
                });
            });
        }
    }

    setTheme(theme) {
        this.currentTheme = theme;
        localStorage.setItem('goclean-theme', theme);
        this.applyTheme();
    }

    applyTheme() {
        const html = document.documentElement;
        const body = document.body;
        
        // Remove existing theme classes
        html.classList.remove('light-theme', 'dark-theme');
        body.classList.remove('light-theme', 'dark-theme');
        
        if (this.currentTheme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            const themeClass = prefersDark ? 'dark-theme' : 'light-theme';
            html.classList.add(themeClass);
            body.classList.add(themeClass);
        } else {
            const themeClass = this.currentTheme + '-theme';
            html.classList.add(themeClass);
            body.classList.add(themeClass);
        }
    }

    toggleTheme() {
        const themes = ['light', 'dark', 'auto'];
        const currentIndex = themes.indexOf(this.currentTheme);
        const nextIndex = (currentIndex + 1) % themes.length;
        this.setTheme(themes[nextIndex]);
    }
}

class EnhancedSearchManager extends ViolationFilter {
    constructor() {
        super();
        this.setupAdvancedSearch();
        this.setupSearchHistory();
    }

    setupAdvancedSearch() {
        const searchContainer = document.querySelector('.search-input-enhanced') || 
                               document.querySelector('.col-md-4');
        
        if (searchContainer && !searchContainer.querySelector('.search-suggestions')) {
            this.createSearchSuggestions(searchContainer);
        }
    }

    createSearchSuggestions(container) {
        const suggestionsDiv = document.createElement('div');
        suggestionsDiv.className = 'search-suggestions position-absolute w-100';
        suggestionsDiv.style.cssText = 'z-index: 1000; top: 100%; display: none;';
        
        const suggestionsList = document.createElement('ul');
        suggestionsList.className = 'list-group shadow';
        suggestionsDiv.appendChild(suggestionsList);
        
        container.style.position = 'relative';
        container.appendChild(suggestionsDiv);

        // Setup suggestion functionality
        this.setupSuggestionLogic(suggestionsList);
    }

    setupSuggestionLogic(suggestionsList) {
        const searchInput = this.searchInput;
        let suggestionTimeout;

        searchInput.addEventListener('input', (e) => {
            clearTimeout(suggestionTimeout);
            suggestionTimeout = setTimeout(() => {
                this.updateSuggestions(suggestionsList, e.target.value);
            }, 200);
        });

        searchInput.addEventListener('focus', () => {
            if (searchInput.value) {
                suggestionsList.parentElement.style.display = 'block';
            }
        });

        document.addEventListener('click', (e) => {
            if (!e.target.closest('.search-suggestions') && !e.target.closest('#searchInput')) {
                suggestionsList.parentElement.style.display = 'none';
            }
        });
    }

    updateSuggestions(suggestionsList, query) {
        if (!query || query.length < 2) {
            suggestionsList.parentElement.style.display = 'none';
            return;
        }

        const suggestions = this.generateSuggestions(query);
        suggestionsList.innerHTML = '';

        suggestions.slice(0, 5).forEach(suggestion => {
            const li = document.createElement('li');
            li.className = 'list-group-item list-group-item-action';
            li.innerHTML = `<i class="bi bi-search me-2"></i>${this.highlightMatch(suggestion.text, query)}`;
            li.addEventListener('click', () => {
                this.searchInput.value = suggestion.value;
                this.applyFilters();
                suggestionsList.parentElement.style.display = 'none';
            });
            suggestionsList.appendChild(li);
        });

        suggestionsList.parentElement.style.display = suggestions.length > 0 ? 'block' : 'none';
    }

    generateSuggestions(query) {
        const suggestions = [];
        const queryLower = query.toLowerCase();

        // File name suggestions
        this.fileItems.forEach(fileItem => {
            const fileName = fileItem.dataset.fileName;
            if (fileName && fileName.toLowerCase().includes(queryLower)) {
                suggestions.push({
                    text: `File: ${fileName}`,
                    value: fileName,
                    type: 'file'
                });
            }
        });

        // Violation type suggestions
        const violationTypes = ['function_length', 'function_complexity', 'naming_convention', 'code_duplication'];
        violationTypes.forEach(type => {
            if (type.toLowerCase().includes(queryLower)) {
                suggestions.push({
                    text: `Type: ${this.formatViolationType(type)}`,
                    value: type,
                    type: 'violation-type'
                });
            }
        });

        return suggestions;
    }

    highlightMatch(text, query) {
        const regex = new RegExp(`(${this.escapeRegExp(query)})`, 'gi');
        return text.replace(regex, '<mark>$1</mark>');
    }

    formatViolationType(type) {
        return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    }

    setupSearchHistory() {
        this.searchHistory = JSON.parse(localStorage.getItem('goclean-search-history') || '[]');
    }

    addToHistory(query) {
        if (query && !this.searchHistory.includes(query)) {
            this.searchHistory.unshift(query);
            this.searchHistory = this.searchHistory.slice(0, 10); // Keep only last 10
            localStorage.setItem('goclean-search-history', JSON.stringify(this.searchHistory));
        }
    }

    focusSearch() {
        this.searchInput.focus();
        this.searchInput.select();
    }

    handleResize() {
        // Adjust search suggestions position on resize
        const suggestions = document.querySelector('.search-suggestions');
        if (suggestions && suggestions.style.display === 'block') {
            suggestions.style.display = 'none';
            setTimeout(() => {
                suggestions.style.display = 'block';
            }, 100);
        }
    }
}

class AnimationManager {
    constructor() {
        this.prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        this.animationQueue = [];
        this.setupIntersectionObserver();
    }

    setupIntersectionObserver() {
        if (this.prefersReducedMotion) return;

        this.observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.animateElement(entry.target);
                    this.observer.unobserve(entry.target);
                }
            });
        }, {
            threshold: 0.1,
            rootMargin: '50px'
        });

        // Observe elements for animation
        this.observeElements();
    }

    observeElements() {
        const elementsToAnimate = [
            '.stats-card',
            '.violation-card',
            '.accordion-item',
            '.card'
        ];

        elementsToAnimate.forEach(selector => {
            document.querySelectorAll(selector).forEach((element, index) => {
                element.style.opacity = '0';
                element.style.transform = 'translateY(20px)';
                element.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
                element.style.transitionDelay = `${index * 0.1}s`;
                this.observer.observe(element);
            });
        });
    }

    animateElement(element) {
        if (this.prefersReducedMotion) {
            element.style.opacity = '1';
            element.style.transform = 'none';
            return;
        }

        element.style.opacity = '1';
        element.style.transform = 'translateY(0)';
    }

    handleResize() {
        // Pause animations during resize to improve performance
        this.pauseAnimations();
        setTimeout(() => this.resumeAnimations(), 150);
    }

    pauseAnimations() {
        document.body.style.setProperty('--transition-fast', '0s');
        document.body.style.setProperty('--transition-normal', '0s');
        document.body.style.setProperty('--transition-slow', '0s');
    }

    resumeAnimations() {
        if (this.prefersReducedMotion) return;
        
        document.body.style.removeProperty('--transition-fast');
        document.body.style.removeProperty('--transition-normal');
        document.body.style.removeProperty('--transition-slow');
    }
}

class AccessibilityManager {
    constructor() {
        this.setupSkipLinks();
        this.setupKeyboardNavigation();
        this.setupAriaLabels();
        this.setupFocusManagement();
    }

    setupSkipLinks() {
        const skipLink = document.createElement('a');
        skipLink.href = '#main-content';
        skipLink.className = 'skip-link sr-only sr-only-focusable';
        skipLink.textContent = 'Skip to main content';
        skipLink.style.cssText = `
            position: absolute;
            left: -10000px;
            top: auto;
            width: 1px;
            height: 1px;
            overflow: hidden;
        `;
        
        skipLink.addEventListener('focus', () => {
            skipLink.style.cssText = `
                position: absolute;
                left: 6px;
                top: 7px;
                z-index: 999999;
                padding: 8px 16px;
                background: #000;
                color: #fff;
                text-decoration: none;
            `;
        });
        
        skipLink.addEventListener('blur', () => {
            skipLink.style.cssText = `
                position: absolute;
                left: -10000px;
                top: auto;
                width: 1px;
                height: 1px;
                overflow: hidden;
            `;
        });

        document.body.insertBefore(skipLink, document.body.firstChild);
    }

    setupKeyboardNavigation() {
        // Add tabindex to interactive elements
        document.querySelectorAll('.violation-card, .accordion-button').forEach(element => {
            if (!element.hasAttribute('tabindex')) {
                element.setAttribute('tabindex', '0');
            }
        });

        // Handle Enter and Space for custom interactive elements
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                const target = e.target;
                if (target.classList.contains('violation-card')) {
                    e.preventDefault();
                    // Expand the parent accordion if collapsed
                    const accordion = target.closest('.accordion-item');
                    if (accordion) {
                        const button = accordion.querySelector('.accordion-button');
                        if (button) button.click();
                    }
                }
            }
        });
    }

    setupAriaLabels() {
        // Add proper aria labels
        document.querySelectorAll('.stats-card').forEach((card, index) => {
            const title = card.querySelector('.card-title')?.textContent;
            const description = card.querySelector('.card-text')?.textContent;
            if (title && description) {
                card.setAttribute('aria-label', `${description}: ${title}`);
            }
        });

        // Add aria-live regions for dynamic content
        const searchResults = document.querySelector('.accordion');
        if (searchResults) {
            searchResults.setAttribute('aria-live', 'polite');
            searchResults.setAttribute('aria-label', 'Search results');
        }
    }

    setupFocusManagement() {
        // Manage focus for modals and overlays
        document.addEventListener('shown.bs.modal', (e) => {
            const modal = e.target;
            const firstFocusable = modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
            if (firstFocusable) {
                firstFocusable.focus();
            }
        });

        // Trap focus in modals
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                const modal = document.querySelector('.modal.show');
                if (modal) {
                    const focusableElements = modal.querySelectorAll(
                        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
                    );
                    const firstElement = focusableElements[0];
                    const lastElement = focusableElements[focusableElements.length - 1];

                    if (e.shiftKey && document.activeElement === firstElement) {
                        e.preventDefault();
                        lastElement.focus();
                    } else if (!e.shiftKey && document.activeElement === lastElement) {
                        e.preventDefault();
                        firstElement.focus();
                    }
                }
            }
        });
    }
}

class PerformanceManager {
    constructor() {
        this.performanceMetrics = {};
        this.isMonitoring = false;
        this.setupPerformanceObserver();
    }

    setupPerformanceObserver() {
        if ('PerformanceObserver' in window) {
            const observer = new PerformanceObserver((list) => {
                list.getEntries().forEach((entry) => {
                    this.recordMetric(entry);
                });
            });
            
            try {
                observer.observe({ entryTypes: ['measure', 'navigation', 'paint'] });
            } catch (e) {
                console.warn('Performance Observer not fully supported');
            }
        }
    }

    startMonitoring() {
        this.isMonitoring = true;
        this.recordLoadTime();
        this.monitorRenderPerformance();
    }

    recordLoadTime() {
        window.addEventListener('load', () => {
            const perfData = performance.getEntriesByType('navigation')[0];
            if (perfData) {
                this.performanceMetrics.loadTime = perfData.loadEventEnd - perfData.loadEventStart;
                this.performanceMetrics.domContentLoaded = perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart;
            }
        });
    }

    monitorRenderPerformance() {
        // Monitor chart rendering performance
        const chartElements = document.querySelectorAll('canvas');
        chartElements.forEach(canvas => {
            const resizeObserver = new ResizeObserver(() => {
                this.debounceChartResize(canvas);
            });
            resizeObserver.observe(canvas);
        });
    }

    debounceChartResize(canvas) {
        clearTimeout(canvas._resizeTimeout);
        canvas._resizeTimeout = setTimeout(() => {
            // Chart.js will handle the resize automatically
            const chart = Chart.getChart(canvas);
            if (chart) {
                chart.resize();
            }
        }, 150);
    }

    recordMetric(entry) {
        if (!this.isMonitoring) return;
        
        this.performanceMetrics[entry.name] = {
            duration: entry.duration,
            timestamp: entry.startTime
        };
    }

    pauseAnimations() {
        document.body.classList.add('animations-paused');
    }

    resumeAnimations() {
        document.body.classList.remove('animations-paused');
    }

    getMetrics() {
        return { ...this.performanceMetrics };
    }
}

class ExportManager {
    constructor() {
        this.setupExportButton();
    }

    setupExportButton() {
        // Create export dropdown in navbar
        const navbar = document.querySelector('.navbar-nav');
        if (navbar && !navbar.querySelector('.export-dropdown')) {
            const exportDropdown = document.createElement('div');
            exportDropdown.className = 'nav-item dropdown export-dropdown';
            exportDropdown.innerHTML = `
                <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                    <i class="bi bi-download"></i> Export
                </a>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li><a class="dropdown-item" href="#" data-export="pdf"><i class="bi bi-file-pdf"></i> Export as PDF</a></li>
                    <li><a class="dropdown-item" href="#" data-export="csv"><i class="bi bi-filetype-csv"></i> Export as CSV</a></li>
                    <li><a class="dropdown-item" href="#" data-export="json"><i class="bi bi-filetype-json"></i> Export as JSON</a></li>
                </ul>
            `;
            navbar.appendChild(exportDropdown);

            // Add event listeners
            exportDropdown.querySelectorAll('[data-export]').forEach(link => {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.exportReport(e.target.closest('[data-export]').dataset.export);
                });
            });
        }
    }

    exportReport(format) {
        switch (format) {
            case 'pdf':
                this.exportToPDF();
                break;
            case 'csv':
                this.exportToCSV();
                break;
            case 'json':
                this.exportToJSON();
                break;
        }
    }

    exportToPDF() {
        // Use browser's print functionality for PDF export
        window.print();
    }

    exportToCSV() {
        const violations = this.extractViolationsData();
        const csv = this.convertToCSV(violations);
        this.downloadFile(csv, 'violations.csv', 'text/csv');
    }

    exportToJSON() {
        const data = {
            summary: this.extractSummaryData(),
            violations: this.extractViolationsData(),
            exportedAt: new Date().toISOString()
        };
        const json = JSON.stringify(data, null, 2);
        this.downloadFile(json, 'violations.json', 'application/json');
    }

    extractSummaryData() {
        const summary = {};
        document.querySelectorAll('.stats-card').forEach((card, index) => {
            const title = card.querySelector('.card-text')?.textContent.trim();
            const value = card.querySelector('.card-title')?.textContent.trim();
            if (title && value) {
                summary[title.toLowerCase().replace(/\s+/g, '_')] = value;
            }
        });
        return summary;
    }

    extractViolationsData() {
        const violations = [];
        document.querySelectorAll('.file-item').forEach(fileItem => {
            const filePath = fileItem.dataset.filePath;
            fileItem.querySelectorAll('.violation-card').forEach(card => {
                violations.push({
                    file: filePath,
                    severity: card.dataset.severity,
                    type: card.dataset.type,
                    line: card.dataset.line,
                    message: card.querySelector('.card-text')?.textContent.trim(),
                    description: card.querySelector('.text-muted')?.textContent.trim()
                });
            });
        });
        return violations;
    }

    convertToCSV(data) {
        if (!data.length) return '';
        
        const headers = Object.keys(data[0]);
        const csvContent = [
            headers.join(','),
            ...data.map(row => 
                headers.map(header => `"${(row[header] || '').toString().replace(/"/g, '""')}"`).join(',')
            )
        ].join('\n');
        
        return csvContent;
    }

    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
    }

    showExportDialog() {
        // Focus on export dropdown if it exists
        const exportToggle = document.querySelector('.export-dropdown .dropdown-toggle');
        if (exportToggle) {
            exportToggle.click();
        }
    }
}

// Initialize enhanced report manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.enhancedReportManager = new EnhancedReportManager();
});

// Quick actions setup
document.addEventListener('DOMContentLoaded', () => {
    const quickActions = document.createElement('div');
    quickActions.className = 'quick-actions d-none d-lg-block';
    quickActions.innerHTML = `
        <button class="quick-action-btn" title="Back to top" onclick="window.scrollTo({top: 0, behavior: 'smooth'})">
            <i class="bi bi-arrow-up"></i>
        </button>
        <button class="quick-action-btn" title="Toggle theme" onclick="window.enhancedReportManager?.themeManager.toggleTheme()">
            <i class="bi bi-palette"></i>
        </button>
        <button class="quick-action-btn" title="Focus search" onclick="window.enhancedReportManager?.searchManager.focusSearch()">
            <i class="bi bi-search"></i>
        </button>
    `;
    document.body.appendChild(quickActions);

    // Show/hide back to top button based on scroll position
    const backToTopBtn = quickActions.querySelector('[title="Back to top"]');
    window.addEventListener('scroll', () => {
        if (window.scrollY > 300) {
            backToTopBtn.style.display = 'flex';
        } else {
            backToTopBtn.style.display = 'none';
        }
    });
    
    // Initially hide the back to top button
    backToTopBtn.style.display = 'none';
});

// CSS for animations-paused class
const style = document.createElement('style');
style.textContent = `
    .animations-paused * {
        animation-duration: 0s !important;
        transition-duration: 0s !important;
    }
`;
document.head.appendChild(style);
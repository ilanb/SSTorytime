/**
 * Tooltip System - Fixed position tooltips that escape overflow:hidden containers
 * Uses data-tooltip attribute for content and data-tooltip-pos for positioning
 */
(function() {
    'use strict';

    // Create tooltip element
    const tooltip = document.createElement('div');
    tooltip.className = 'js-tooltip';
    tooltip.setAttribute('role', 'tooltip');
    document.body.appendChild(tooltip);

    // Create arrow element
    const arrow = document.createElement('div');
    arrow.className = 'js-tooltip-arrow';
    tooltip.appendChild(arrow);

    // Create content element
    const content = document.createElement('div');
    content.className = 'js-tooltip-content';
    tooltip.appendChild(content);

    let currentTarget = null;
    let showTimeout = null;
    let hideTimeout = null;

    // Position the tooltip relative to target element
    function positionTooltip(target) {
        const rect = target.getBoundingClientRect();
        const pos = target.getAttribute('data-tooltip-pos') || 'top';
        const tooltipRect = tooltip.getBoundingClientRect();

        let top, left;
        const gap = 8;
        const arrowSize = 6;

        // Reset arrow classes
        arrow.className = 'js-tooltip-arrow';

        switch (pos) {
            case 'bottom':
                top = rect.bottom + gap;
                left = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
                arrow.classList.add('arrow-top');
                break;
            case 'left':
                top = rect.top + (rect.height / 2) - (tooltipRect.height / 2);
                left = rect.left - tooltipRect.width - gap;
                arrow.classList.add('arrow-right');
                break;
            case 'right':
                top = rect.top + (rect.height / 2) - (tooltipRect.height / 2);
                left = rect.right + gap;
                arrow.classList.add('arrow-left');
                break;
            case 'top':
            default:
                top = rect.top - tooltipRect.height - gap;
                left = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
                arrow.classList.add('arrow-bottom');
                break;
        }

        // Keep tooltip within viewport
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        const margin = 10;

        // Horizontal bounds
        if (left < margin) {
            left = margin;
        } else if (left + tooltipRect.width > viewportWidth - margin) {
            left = viewportWidth - tooltipRect.width - margin;
        }

        // Vertical bounds - flip if needed
        if (top < margin && pos === 'top') {
            // Flip to bottom
            top = rect.bottom + gap;
            arrow.className = 'js-tooltip-arrow arrow-top';
        } else if (top + tooltipRect.height > viewportHeight - margin && pos === 'bottom') {
            // Flip to top
            top = rect.top - tooltipRect.height - gap;
            arrow.className = 'js-tooltip-arrow arrow-bottom';
        }

        tooltip.style.top = `${top}px`;
        tooltip.style.left = `${left}px`;

        // Position arrow horizontally centered on target
        const arrowLeft = rect.left + (rect.width / 2) - left - arrowSize;
        arrow.style.left = `${Math.max(arrowSize, Math.min(arrowLeft, tooltipRect.width - arrowSize * 3))}px`;
    }

    function showTooltip(target) {
        const text = target.getAttribute('data-tooltip');
        if (!text) return;

        currentTarget = target;
        content.textContent = text;

        // Make visible but transparent to measure
        tooltip.style.visibility = 'hidden';
        tooltip.style.opacity = '0';
        tooltip.classList.add('visible');

        // Position after render
        requestAnimationFrame(() => {
            positionTooltip(target);
            tooltip.style.visibility = 'visible';
            tooltip.style.opacity = '1';
        });
    }

    function hideTooltip() {
        tooltip.classList.remove('visible');
        tooltip.style.opacity = '0';
        currentTarget = null;
    }

    // Event delegation for better performance
    document.addEventListener('mouseenter', function(e) {
        // e.target might be a text node or other non-element node
        if (!e.target || !e.target.closest) return;

        const target = e.target.closest('[data-tooltip]');
        if (!target) return;

        clearTimeout(hideTimeout);
        showTimeout = setTimeout(() => showTooltip(target), 50);
    }, true);

    document.addEventListener('mouseleave', function(e) {
        // e.target might be a text node or other non-element node
        if (!e.target || !e.target.closest) return;

        const target = e.target.closest('[data-tooltip]');
        if (!target) return;

        clearTimeout(showTimeout);
        hideTimeout = setTimeout(hideTooltip, 100);
    }, true);

    // Hide on scroll
    document.addEventListener('scroll', function() {
        if (currentTarget) {
            hideTooltip();
        }
    }, true);

    // Reposition on window resize
    window.addEventListener('resize', function() {
        if (currentTarget) {
            positionTooltip(currentTarget);
        }
    });

    // Hide on escape key
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && currentTarget) {
            hideTooltip();
        }
    });
})();

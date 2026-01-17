import React, { useRef, useEffect, useCallback } from 'react';

interface TouchPoint {
  x: number;
  y: number;
  timestamp: number;
  type: 'down' | 'up' | 'move';
}

interface TouchOverlayProps {
  width: number;
  height: number;
  videoWidth: number;
  videoHeight: number;
  touchEvents: TouchPoint[];
  currentTimeMs: number;
  showTrail?: boolean;
  trailDurationMs?: number;
}

const TouchOverlay: React.FC<TouchOverlayProps> = ({
  width,
  height,
  videoWidth,
  videoHeight,
  touchEvents,
  currentTimeMs,
  showTrail = true,
  trailDurationMs = 500,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Calculate scale factors
  const scaleX = width / videoWidth;
  const scaleY = height / videoHeight;

  const drawTouchPoints = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    // Filter events within the trail window
    const trailStart = currentTimeMs - trailDurationMs;
    const relevantEvents = touchEvents.filter(
      (e) => e.timestamp >= trailStart && e.timestamp <= currentTimeMs
    );

    if (relevantEvents.length === 0) return;

    // Draw trail
    if (showTrail && relevantEvents.length > 1) {
      ctx.beginPath();
      ctx.strokeStyle = 'rgba(255, 100, 100, 0.6)';
      ctx.lineWidth = 3;
      ctx.lineCap = 'round';
      ctx.lineJoin = 'round';

      let started = false;
      for (const event of relevantEvents) {
        const x = event.x * scaleX;
        const y = event.y * scaleY;

        if (!started) {
          ctx.moveTo(x, y);
          started = true;
        } else {
          ctx.lineTo(x, y);
        }
      }
      ctx.stroke();
    }

    // Draw current touch point
    const currentEvent = relevantEvents[relevantEvents.length - 1];
    if (currentEvent && currentTimeMs - currentEvent.timestamp < 200) {
      const x = currentEvent.x * scaleX;
      const y = currentEvent.y * scaleY;

      // Outer glow
      const gradient = ctx.createRadialGradient(x, y, 0, x, y, 30);
      gradient.addColorStop(0, 'rgba(255, 100, 100, 0.8)');
      gradient.addColorStop(0.5, 'rgba(255, 100, 100, 0.3)');
      gradient.addColorStop(1, 'rgba(255, 100, 100, 0)');

      ctx.beginPath();
      ctx.fillStyle = gradient;
      ctx.arc(x, y, 30, 0, Math.PI * 2);
      ctx.fill();

      // Inner circle
      ctx.beginPath();
      ctx.fillStyle = 'rgba(255, 255, 255, 0.9)';
      ctx.arc(x, y, 8, 0, Math.PI * 2);
      ctx.fill();

      // Border
      ctx.beginPath();
      ctx.strokeStyle = 'rgba(255, 100, 100, 1)';
      ctx.lineWidth = 2;
      ctx.arc(x, y, 8, 0, Math.PI * 2);
      ctx.stroke();

      // Touch type indicator
      if (currentEvent.type === 'down') {
        ctx.beginPath();
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
        ctx.lineWidth = 2;
        ctx.arc(x, y, 15, 0, Math.PI * 2);
        ctx.stroke();
      }
    }
  }, [width, height, scaleX, scaleY, touchEvents, currentTimeMs, showTrail, trailDurationMs]);

  useEffect(() => {
    drawTouchPoints();
  }, [drawTouchPoints]);

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      style={{
        position: 'absolute',
        top: 0,
        left: 0,
        pointerEvents: 'none',
        zIndex: 10,
      }}
    />
  );
};

export default TouchOverlay;

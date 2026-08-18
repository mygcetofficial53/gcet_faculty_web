"use client";

import { motion } from "framer-motion";
import { Book, GraduationCap, Microscope, Code, Calculator, Atom, Library, FlaskConical } from "lucide-react";
import { useEffect, useState, useMemo } from "react";

const ICONS = [Book, GraduationCap, Microscope, Code, Calculator, Atom, Library, FlaskConical];

export function FloatingBackground() {
  const [windowSize, setWindowSize] = useState({ width: 0, height: 0 });

  useEffect(() => {
    // Only access window on client
    setWindowSize({ width: window.innerWidth, height: window.innerHeight });
    
    const handleResize = () => {
      setWindowSize({ width: window.innerWidth, height: window.innerHeight });
    };
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const randomValues = useMemo(() => {
    return Array.from({ length: 15 }).map((_, i) => ({
      Icon: ICONS[i % ICONS.length],
      startX: Math.random() * windowSize.width,
      startY: Math.random() * windowSize.height,
      driftX: (Math.random() - 0.5) * 200,
      driftY: (Math.random() - 0.5) * 200,
      initialRotate: Math.random() * 360,
      animateRotate: Math.random() * 360 + 180,
      duration: 20 + Math.random() * 20,
      size: 40 + Math.random() * 60,
    }));
  }, [windowSize.width, windowSize.height]);

  if (windowSize.width === 0) return null;

  return (
    <div className="fixed inset-0 overflow-hidden pointer-events-none -z-10 bg-academic-grid">
      {/* Subtle overlay to soften the grid */}
      <div className="absolute inset-0 bg-background/60 backdrop-blur-[1px]"></div>
      
      {/* Animated Icons */}
      {randomValues.map((item, i) => {
        const { Icon, startX, startY, driftX, driftY, initialRotate, animateRotate, duration, size } = item;
        
        return (
          <motion.div
            key={i}
            initial={{ opacity: 0, x: startX, y: startY, rotate: initialRotate }}
            animate={{
              opacity: [0.1, 0.4, 0.1],
              x: startX + driftX,
              y: startY + driftY,
              rotate: animateRotate,
            }}
            transition={{
              duration: duration,
              repeat: Infinity,
              repeatType: "reverse",
              ease: "linear",
            }}
            className="absolute text-primary/10"
          >
            <Icon size={size} strokeWidth={1} />
          </motion.div>
        );
      })}
    </div>
  );
}

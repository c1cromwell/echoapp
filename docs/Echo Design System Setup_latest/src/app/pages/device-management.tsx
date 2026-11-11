import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, Smartphone, Laptop, Tablet, MonitorSmartphone } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { StepUpModal } from "../components/step-up-modal";

interface Device {
  id: string;
  name: string;
  type: "phone" | "tablet" | "laptop";
  os: string;
  lastActive: string;
  location: string;
  isCurrent: boolean;
}

const mockDevices: Device[] = [
  {
    id: "d1",
    name: "iPhone 15 Pro",
    type: "phone",
    os: "iOS 17.4",
    lastActive: "Active now",
    location: "New York, NY",
    isCurrent: true,
  },
  {
    id: "d2",
    name: "iPad Air",
    type: "tablet",
    os: "iPadOS 17.2",
    lastActive: "2 hrs ago",
    location: "New York, NY",
    isCurrent: false,
  },
  {
    id: "d3",
    name: "MacBook Pro",
    type: "laptop",
    os: "macOS 14.3",
    lastActive: "3 days ago",
    location: "Brooklyn, NY",
    isCurrent: false,
  },
];

const deviceIcons: Record<string, React.ElementType> = {
  phone: Smartphone,
  tablet: Tablet,
  laptop: Laptop,
};

export function DeviceManagementPage() {
  const navigate = useNavigate();
  const [devices, setDevices] = useState(mockDevices);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [pendingRemoveId, setPendingRemoveId] = useState<string | null>(null);
  const [removedDevice, setRemovedDevice] = useState<Device | null>(null);
  const [showLogoutAll, setShowLogoutAll] = useState(false);

  const currentDevice = devices.find((d) => d.isCurrent);
  const otherDevices = devices.filter((d) => !d.isCurrent);

  const handleRemove = (id: string) => {
    setPendingRemoveId(id);
    setStepUpOpen(true);
  };

  const handleVerified = () => {
    setStepUpOpen(false);
    if (pendingRemoveId) {
      const device = devices.find((d) => d.id === pendingRemoveId);
      if (device) setRemovedDevice(device);
      setDevices((prev) => prev.filter((d) => d.id !== pendingRemoveId));
      setPendingRemoveId(null);
    }
    if (showLogoutAll) {
      setDevices((prev) => prev.filter((d) => d.isCurrent));
      setShowLogoutAll(false);
    }
  };

  const handleLogoutAll = () => {
    setShowLogoutAll(true);
    setStepUpOpen(true);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-3 px-4 pt-14 pb-3">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
        <h1
          style={{
            fontSize: 20,
            fontWeight: 600,
            color: "var(--echo-text-primary)",
            margin: 0,
          }}
        >
          Devices
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-6">
        {/* This Device */}
        {currentDevice && (
          <div className="mb-5">
            <h3
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: "var(--echo-text-secondary)",
                margin: "0 0 10px 4px",
              }}
            >
              This Device
            </h3>
            <DeviceCard device={currentDevice} />
          </div>
        )}

        {/* Other Devices */}
        <div className="mb-5">
          <h3
            style={{
              fontSize: 14,
              fontWeight: 600,
              color: "var(--echo-text-secondary)",
              margin: "0 0 10px 4px",
            }}
          >
            Other Devices
          </h3>

          {otherDevices.length > 0 ? (
            <div className="flex flex-col gap-3">
              {otherDevices.map((device) => (
                <DeviceCard
                  key={device.id}
                  device={device}
                  onRemove={() => handleRemove(device.id)}
                />
              ))}
            </div>
          ) : (
            <div
              className="flex flex-col items-center gap-3 rounded-2xl py-8"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                border: "1px solid var(--echo-border)",
              }}
            >
              <MonitorSmartphone size={40} style={{ color: "var(--echo-text-tertiary)" }} />
              <div className="text-center px-6">
                <p style={{ fontSize: 16, fontWeight: 500, color: "var(--echo-text-secondary)", margin: "0 0 4px" }}>
                  No other devices registered
                </p>
                <p style={{ fontSize: 14, color: "var(--echo-text-tertiary)", margin: 0, lineHeight: "20px" }}>
                  You can add devices by scanning a QR code from your primary device.
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Divider + Log Out All */}
        {otherDevices.length > 0 && (
          <>
            <div style={{ height: 1, backgroundColor: "var(--echo-border)", margin: "0 0 16px" }} />
            <button
              onClick={handleLogoutAll}
              className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center active:scale-[0.98] transition-transform"
              style={{
                height: 48,
                backgroundColor: "var(--echo-error)",
                fontSize: 16,
                fontWeight: 600,
                color: "white",
                borderRadius: 12,
              }}
            >
              Log Out All Devices
            </button>
          </>
        )}
      </div>

      {/* Step-Up Modal */}
      <StepUpModal
        isOpen={stepUpOpen}
        onClose={() => {
          setStepUpOpen(false);
          setPendingRemoveId(null);
          setShowLogoutAll(false);
        }}
        onVerified={handleVerified}
        subtitle={
          showLogoutAll
            ? "Logging out all other devices requires verification."
            : "Removing a device requires your verification."
        }
      />

      {/* Removed confirmation overlay */}
      <AnimatePresence>
        {removedDevice && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center px-10"
            style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
          >
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="w-full rounded-[20px] p-6 text-center"
              style={{
                maxWidth: 320,
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
              }}
            >
              <h3 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 8px" }}>
                Device removed
              </h3>
              <p style={{ fontSize: 16, color: "var(--echo-text-secondary)", margin: "0 0 20px", lineHeight: "24px" }}>
                {removedDevice.name} has been logged out and can no longer access your account.
              </p>
              <button
                onClick={() => setRemovedDevice(null)}
                className="w-full border-none cursor-pointer rounded-xl"
                style={{
                  height: 48,
                  backgroundColor: "var(--echo-primary)",
                  fontSize: 16,
                  fontWeight: 600,
                  color: "white",
                  borderRadius: 12,
                }}
              >
                Done
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

function DeviceCard({
  device,
  onRemove,
}: {
  device: Device;
  onRemove?: () => void;
}) {
  const Icon = deviceIcons[device.type] || Smartphone;

  return (
    <div
      className="rounded-2xl px-4 py-3.5"
      style={{
        backgroundColor: "var(--echo-bg-elevated)",
        border: "1px solid var(--echo-border)",
        boxShadow: "var(--echo-shadow-sm)",
      }}
    >
      <div className="flex items-start gap-3">
        <div
          className="flex items-center justify-center rounded-xl shrink-0 mt-0.5"
          style={{
            width: 40,
            height: 40,
            backgroundColor: "var(--echo-primary-subtle)",
          }}
        >
          <Icon size={20} style={{ color: "var(--echo-primary)" }} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
              {device.name}
            </span>
            {device.isCurrent && (
              <div
                className="rounded-full"
                style={{ width: 8, height: 8, backgroundColor: "var(--echo-success)" }}
              />
            )}
          </div>
          <div style={{ fontSize: 14, color: "var(--echo-text-secondary)", marginTop: 2 }}>
            {device.os} {" \u00B7 "} {device.lastActive}
          </div>
          <div className="flex items-center justify-between mt-1">
            <span style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>
              {device.location}
            </span>
            {onRemove && (
              <button
                onClick={onRemove}
                className="bg-transparent border-none cursor-pointer p-0"
                style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-error)" }}
              >
                Remove
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
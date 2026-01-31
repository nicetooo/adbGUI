import React from "react";
import { Button, Result } from "antd";

interface Props {
  children: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error("[ErrorBoundary]", error, info.componentStack);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <Result
          status="error"
          title="Something went wrong"
          subTitle={this.state.error?.message || "An unexpected error occurred"}
          extra={[
            <Button key="retry" type="primary" onClick={this.handleRetry}>
              Retry
            </Button>,
            <Button key="reload" onClick={this.handleReload}>
              Reload App
            </Button>,
          ]}
        >
          <pre
            style={{
              maxHeight: 200,
              overflow: "auto",
              fontSize: 12,
              textAlign: "left",
              background: "var(--ant-color-bg-layout)",
              padding: 12,
              borderRadius: 6,
            }}
          >
            {this.state.error?.stack}
          </pre>
        </Result>
      );
    }
    return this.props.children;
  }
}

export default ErrorBoundary;

function App() {
    const [address, setAddress] = React.useState('');
    const [isValid, setIsValid] = React.useState(true);

    const handleSubmit = (e) => {
        e.preventDefault();
        if (address.length === 44) {
            console.log('Valid Solana address submitted:', address);
            // Add your submission logic here
        } else {
            setIsValid(false);
        }
    };

    const createFloatingObjects = () => {
        const objects = [];
        const colors = ['var(--neon-blue)', 'var(--deep-purple)', 'var(--bright-pink)'];
        const shapes = ['circle', 'triangle', 'hexagon'];
        
        for (let i = 0; i < 20; i++) {
            const shape = shapes[Math.floor(Math.random() * shapes.length)];
            const style = {
                left: `${Math.random() * 100}%`,
                top: `${Math.random() * 100}%`,
                width: `${Math.random() * 50 + 20}px`,
                height: `${Math.random() * 50 + 20}px`,
                backgroundColor: colors[Math.floor(Math.random() * colors.length)],
                animationDelay: `${Math.random() * 5}s`,
                animationDuration: `${Math.random() * 10 + 5}s`,
                filter: `blur(${Math.random() * 2}px)`,
                opacity: Math.random() * 0.5 + 0.3,
            };
            objects.push(<div key={i} className={`floating-object ${shape}`} style={style} />);
        }
        return objects;
    };

    return (
        <div className="container">
            <div className="cyber-grid"></div>
            <div className="floating-objects">
                {createFloatingObjects()}
            </div>
            <div className="glitch-container">
                <h1 className="title glitch" data-text="MAKIMO.LIVE">MAKIMO.LIVE</h1>
            </div>
            <div className="form-container glass-morphism">
                <h2 className="subtitle">Create Your AI VTuber</h2>
                <form onSubmit={handleSubmit} className="cyber-form">
                    <div className="input-group">
                        <input
                            type="text"
                            className="cyber-input"
                            placeholder="Enter Solana Address"
                            value={address}
                            onChange={(e) => {
                                setAddress(e.target.value);
                                setIsValid(true);
                            }}
                            style={{
                                borderColor: isValid ? 'var(--cyber-blue)' : 'var(--neon-pink)'
                            }}
                        />
                        {!isValid && (
                            <p className="error-message">
                                Please enter a valid Solana address Base 58 keypair
                            </p>
                        )}
                    </div>
                    <button type="submit" className="submit-btn">
                        <span className="btn-text">Generate VTuber</span>
                        <div className="btn-glow"></div>
                    </button>
                </form>
            </div>
        </div>
    );
}

ReactDOM.render(<App />, document.getElementById('root')); 

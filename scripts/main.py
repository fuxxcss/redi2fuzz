# main.py
import logging
import sys
from .orchestrator import Orchestrator
from .config import DEFAULT_CONFIG

def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.StreamHandler(sys.stdout)
        ]
    )

def main():
    setup_logging()
    logger = logging.getLogger(__name__)
    
    # 创建Orchestrator实例
    orchestrator = Orchestrator()
    
    # 初始化
    if not orchestrator.initialize(DEFAULT_CONFIG):
        logger.error("Failed to initialize orchestrator")
        return 1
    
    # 运行示例分析
    logger.info("Running SAST analysis...")
    sast_result = orchestrator.run_analysis("example_binary", "sast")
    print(f"SAST Result: {sast_result}")
    
    logger.info("Running Fuzz analysis...")
    fuzz_result = orchestrator.run_analysis("example_binary", "fuzz")
    print(f"Fuzz Result: {fuzz_result}")
    
    return 0

if __name__ == "__main__":
    sys.exit(main())
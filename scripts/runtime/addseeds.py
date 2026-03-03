# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description="向AFL++模糊测试活动添加新种子文件",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
使用示例:
  %(prog)s -o /path/to/afl-out-dir /path/to/seed1 /path/to/seed2
  %(prog)s -o /path/to/afl-out-dir -r /path/to/seeds/
  %(prog)s -o /path/to/afl-out-dir -i seed_list.txt
  
与种子池管理包集成:
  可以通过导入Corpus类，在Python代码中直接使用:
  
  from afl_addseeds import Corpus
  
  manager = Corpus("/path/to/afl-out-dir")
  added_files = manager.add_seeds(["/path/to/seed1", "/path/to/seed2"])
        """
    )
    
    parser.add_argument(
        "-o", "--output-dir",
        required=True,
        help="AFL++输出目录（使用'afl-fuzz -o'指定的目录）"
    )
    
    parser.add_argument(
        "-i", "--input-file",
        help="包含种子文件路径列表的文本文件（每行一个路径）"
    )
    
    parser.add_argument(
        "-r", "--recursive",
        action="store_true",
        help="递归处理目录中的文件"
    )
    
    parser.add_argument(
        "-v", "--verbose",
        action="store_true",
        help="显示详细输出"
    )
    
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="模拟运行，不实际复制文件"
    )
    
    parser.add_argument(
        "seed_paths",
        nargs="*",
        help="种子文件或目录的路径"
    )
    
    return parser.parse_args()


def main():
    """主函数"""
    args = parse_arguments()
    
    # 设置日志级别
    if args.verbose:
        logger.setLevel(logging.DEBUG)
    
    try:
        # 创建种子管理器
        if args.dry_run:
            logger.info(f"[模拟运行] 将处理输出目录: {args.output_dir}")
            seed_paths = []
        else:
            manager = Corpus(args.output_dir)
            logger.info(f"已初始化AFL种子管理器，输出目录: {args.output_dir}")
            logger.info(f"种子目录: {manager.seed_dir}")
            
            # 收集种子路径
            seed_paths = list(args.seed_paths)
            
            # 如果指定了输入文件，读取其中的路径
            if args.input_file:
                input_file = Path(args.input_file)
                if input_file.exists():
                    with open(input_file, 'r') as f:
                        file_paths = [line.strip() for line in f if line.strip()]
                        seed_paths.extend(file_paths)
            
            if not seed_paths:
                logger.error("错误: 未指定种子文件或目录")
                sys.exit(1)
            
            # 添加种子
            logger.info(f"开始添加种子... (共{len(seed_paths)}个路径)")
            added_files = manager.add_seeds(seed_paths, args.recursive)
            
            # 显示结果
            logger.info(f"完成! 成功添加了 {len(added_files)} 个种子文件")
            logger.info(f"种子目录中现有文件总数: {manager.get_seed_count()}")
            
            # 如果启用了详细模式，显示添加的文件列表
            if args.verbose and added_files:
                logger.info("添加的文件:")
                for file in added_files:
                    logger.info(f"  {file}")
    
    except ValueError as e:
        logger.error(f"错误: {e}")
        sys.exit(1)
    except Exception as e:
        logger.error(f"意外错误: {e}")
        if args.verbose:
            import traceback
            traceback.print_exc()
        sys.exit(1)
<?php
$texts = array(
	'最近你还好吗？是的发送到发送到水电费水电费水电费水电费水电费水电费多少\sdfsdf',
	'Reading ends when length - 1 bytes have been read, or a newline (which is included in the return value), or an EOF (whichever comes first). If no length is specified, it will keep reading from the stream until it reaches the end of the line. ',
	'       今天我也不知道写个什么帖子了，前期没有怎么准备，没有去探店，现在也没有什么好的素材，所以今天只能写个心情日记了。大斌去过八月照相馆那面两次了，一次是缴剩余的尾款去了，还一次是因为同学的要去八月改单，我也正好和老婆一起去珂兰钻戒那里测量戒指去，就这样和同学一起去了。虽然去了两次，但是这两次我也没有怎么拍照，当时也没有这个心情拍照，主要是为了同学改套餐这个事情。',
	'        今天晚上回来后，在家里上了上了会网，把今天的第一片原创帖修改完，发了出去，正好老婆也马上到家了，就赶紧去车站接老婆去，我还没有到车站，老婆就已经下车了，我们在回来的路上随便买了份凉皮，因为我回家的时候就饿了，自己在老婆没有回家的时候吃了2个在外面买回来的热馒头哦。',
);

$re = new RtextExport('127.0.0.1');
foreach($texts as $text) {
	$textId = $re->write($text);
	var_dump($textId);
}


//var_dump($ret);

class RtextExport {
	private $handler;
	function __construct($host, $port = RtextHandler::DEFAULT_PORT)
	{
		$this->handler = RtextHandler::getInstance($host, $port);
	}

	function get($textId)
	{
		$ret = $this->handler->doRequest('/text/get', array('textId' => $textId));
		if ($ret['status'] == RtextHandler::STATUS_NOT_FOUND) {
			return false;
		}
		return $ret['text'];
	}

	// 写入文本
	function write($text)
	{
		$textId = RtextHandler::getTextId($text);

		$ret = $this->handler->doRequest('/text/exists', array('textId' => $textId));
		if ($ret['status'] != RtextHandler::STATUS_NOT_FOUND) {
			$data = $this->handler->packText($text);
			var_dump($data);
			$ret = $this->handler->doRequest('/text/write', $data);
			if ($ret['status'] != RtextHandler::STATUS_YES) {
				throw new Exception('rtext.writeFailed');
			}
		}
		return $textId;
	}
}


class RtextHandler {
	private $fp = null;
	const STATUS_NOT_FOUND = 4;
	const STATUS_NO = 3;
	const STATUS_YES = 2;
	const STATUS_ERROR = 1;

	const MIN_COMPRESS_LEN = 512;
	const MAX_VALUE_LENGTH = 512;
	const DEFAULT_PORT = 8890;

	private function __construct()
	{
		// TODO
	}

	private static $handlers;

	public static function getInstance($host, $port)
	{
		$hash = md5($host.'_'.$port);
		if (!isset(self::$handlers[$hash])) {
			$handler = new RtextHandler();
			$handler->init($host, $port);
			self::$handlers[$hash] = $handler;
		}
		return self::$handlers[$hash];
	}

	public function init($host, $port) {
		if ($this->fp) {
			return;
		}
		$this->fp = fsockopen($host, $port, $errno, $errstr, 30);
		if (!$this->fp) {
			throw new Exception("doRequest.u_socketConnectFailed:$errstr {$errno}");
		}
	}

	public function close() {
		if ($this->fp) {
			fwrite($this->fp, "\1");
			fclose($this->fp);
		}
	}

	public function __destruct() {
		$this->close();
	}

	// 封装文本
	public function packText($text) {
		$len = strlen($text);
		$textId = self::getTextId($text);
		
		$flag = 0;
		if ($len > self::MIN_COMPRESS_LEN) {
			$text = gzcompress($text, 7);
			$flag = 1;
		}
		return array(
			'textId' => $textId,
			'flag' => $flag,
			'text' => $text,
		);
	}

	// 获取文本对应的ID
	// $text 原始文本
	// 返回64位ID
	public static function getTextId($text) {
		$hash = md5 ( $text, true );
		$high = substr ( $hash, 0, 8 );
		$low = substr ( $hash, 8, 8 );
		$sign = $high ^ $low;
		$sign1 = hexdec ( bin2hex ( substr ( $sign, 0, 4 ) ) );
		$sign2 = hexdec ( bin2hex ( substr ( $sign, 4, 4 ) ) );
		$id = ($sign1 << 32) | $sign2;
		if ($id < 0) {
			return sprintf('%u', $id);
		}
		return $id;
	}


	// 发送请求
	public function doRequest($url, $data) {
		fwrite($this->fp, $this->packRequest($url, $data));
		$response = false;
		if( !feof($this->fp) ) {
			//fgets Reading ends when length - 1 bytes have been read
			$tmp = fgets($this->fp, 3);
			$length = array_shift(unpack('n*',$tmp));
			$response = $this->unpackResponse(fgets($this->fp, $length + 1));
		}	
		var_dump($url, $response);
		return $response;
	}

	private function unpackResponse($str) {
		$ret = array();
		$ret['status'] = array_shift(unpack('C', substr($str, 0, 1)));
		$ret['data'] = array();
		$str = substr($str, 1);
		while ($str) {
			if (substr($str, 0, 1) == "\0") {
				break;
			}
			$pos = strpos($str, "\0");
			if ($pos > -1) {
				$key = substr($str, 0, $pos);
				$str = substr($str, $pos + 1);

				$len = array_shift(unpack('n*', substr($str, 0, 2)));

				$value = substr($str, 2, $len);
				$ret['data'][$key] = $value;
				$str = substr($str, 2 + $len);
			}
		}
		return $ret;
	}

	private function packRequest($url, array $data) {
		return $url."\0". $this->packData($data);
	}

	private function packData(array $data) {
		$ret = array();
		foreach($data as $key => $value) {
			$len = strlen($value);
			if ($len > self::MAX_VALUE_LENGTH) {
				throw new Exception('packData.u_outMaxRequestLen');
			}
			$ret[] = $key."\0".pack('n', $len).$value;
		}
		return implode('', $ret);
	}
}


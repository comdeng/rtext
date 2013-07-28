<?php
error_reporting(E_ERROR | E_WARNING | E_PARSE | E_NOTICE);

$conn = mysql_connect('192.168.0.188:3307', 'root', 'root');
mysql_select_db('text');
mysql_set_charset('utf8');

define('FETCH_ROW', 100000);


$minId = 0;
$counter = 0;
$re = new RtextExport('127.0.0.1');
while(true) {
	$sql = "select * from t_text0 where text_id> {$minId} order by text_id asc limit ".FETCH_ROW;
	$rs = mysql_query($sql);
	$count = 0;
	$start = microtime(true);
	while($row = mysql_fetch_array($rs)) {
		$count++;
		if ($row['text_zip_flag'] == 1) {
			$text = gzuncompress(substr($row['text_content'], 4));
		} else {
			$text = $row['text_content'];
		}
		$minId = $row['text_id'];
		$counter++;
		//var_dump($minId);
		$re->write($text);
	}
	if ($count < FETCH_ROW) {
		break;
	}
	$end = microtime(true);
	printf("time:%.2fms\tnum:%d\n", $end - $start, $counter);
}

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
		if ($ret['status'] == RtextHandler::STATUS_ERROR) {
			throw new  Exception('rtext.writeError');
		}
		if ($ret['status'] == RtextHandler::STATUS_NOT_FOUND) {
			$data = $this->handler->packText($text);
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
	const MAX_VALUE_LENGTH = 262144; // 256x1024
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
		$data = $this->packRequest($url, $data);
		$len = @fwrite($this->fp, $data);
		if (!$len) {
			throw new Exception('rtext.fwriteFailed');
		}
		$response = false;
		if( !feof($this->fp) ) {
			//fgets Reading ends when length - 1 bytes have been read
			$tmp = fgets($this->fp, 3);
			$length = unpack('n*',$tmp);
			$length = array_shift($length);
			$response = $this->unpackResponse(fgets($this->fp, $length + 1));
		}	
		//var_dump($url, $response);
		return $response;
	}

	private function unpackResponse($str) {
		$ret = array();
		$status = unpack('C*', substr($str, 0, 1));
		$ret['status'] = array_shift($status);
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

				$len = unpack('n*', substr($str, 0, 2));
				$len = array_shift($len);

				$value = substr($str, 2, $len);
				$ret['data'][$key] = $value;
				$str = substr($str, 2 + $len);
			}
		}
		return $ret;
	}

	private function packRequest($url, array $data) {
		$ret = $url."\0". $this->packData($data);
		$len = strlen($ret);
		//var_dump("dataLen:".$len);
		// 首位为\1表示要终端输入
		$ret = "\0".pack('C', $len >> 16).pack('C', ($len >> 8) & 0xFF).pack('C', $len & 0xFF). $ret;
		//var_dump($ret);
		return $ret;
	}

	private function packData(array $data) {
		//var_dump($data);
		$ret = array();
		foreach($data as $key => $value) {
			$len = strlen($value);
			//var_dump('key='.$key.';key length is'.$len);
			if ($len > self::MAX_VALUE_LENGTH) {
				//var_dump($data);
				throw new Exception('packData.u_outMaxRequestLen '.$len);
			}
			// 长度保留最长为24bit(3个字节)。
			$len = pack('C', $len >> 16) . pack('C', ($len >> 8) & 0xFF) . pack('C', $len & 0xFF);
			$ret[] = $key."\0".$len.$value;
		}
		return implode('', $ret);
	}
}

